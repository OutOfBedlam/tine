package engine

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/BurntSushi/toml"
	"github.com/OutOfBedlam/tine/util"
)

type Pipeline struct {
	PipelineConfig
	inputs       []*InletHandler
	outputs      []*OutletHandler
	flows        []*FlowHandler
	ctx          *Context
	logger       *slog.Logger
	plumbingOnce sync.Once
	stopOnce     sync.Once
}

type Option func(*Pipeline) error

var pipelineSerial int64

func WithConfig(conf string) Option {
	return func(p *Pipeline) error {
		if _, err := toml.Decode(conf, p); err != nil {
			return err
		}
		if p.Name == "" {
			serial := atomic.AddInt64(&pipelineSerial, 1)
			p.Name = fmt.Sprintf("pipeline-%d", serial)
		}
		return nil
	}
}

func WithConfigFile(path string) Option {
	return func(p *Pipeline) error {
		if content, err := os.ReadFile(path); err != nil {
			return err
		} else {
			if _, err := toml.Decode(string(content), p); err != nil {
				return err
			}
			if p.Name == "" {
				nameCandidate := filepath.Base(path)
				p.Name = nameCandidate
			}
			return nil
		}
	}
}

func WithName(name string) Option {
	return func(p *Pipeline) error {
		p.Name = name
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(p *Pipeline) error {
		p.logger = logger
		return nil
	}
}

func New(opts ...Option) (*Pipeline, error) {
	p := &Pipeline{}
	// load default config
	if _, err := toml.Decode(DefaultConfigString, &p); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	// init logging
	if p.logger == nil {
		p.logger = util.NewLogger(p.Log)
	}

	// context
	p.ctx = newContext(p).WithLogger(p.logger)
	p.flows = []*FlowHandler{NewFlowHandler(FanInFlow(p.ctx))}
	return p, nil
}

func (p *Pipeline) logMsg(msg string) string {
	if p == nil {
		return fmt.Sprintf("pipeline %s", msg)
	} else {
		return fmt.Sprintf("pipeline %s %s", p.Name, msg)
	}
}

func (p *Pipeline) Context() *Context {
	return p.ctx
}

func (p *Pipeline) AddInlet(name string, inlet Inlet) (*InletHandler, error) {
	ret, err := NewInletHandler(name, inlet, p.flows[0].inCh)
	if err != nil {
		return nil, err
	}
	p.inputs = append(p.inputs, ret)
	return ret, nil
}

func (p *Pipeline) AddOutlet(outlet Outlet) (*OutletHandler, error) {
	ret, err := NewOutletHandler(outlet)
	if err != nil {
		return nil, err
	}
	p.outputs = append(p.outputs, ret)
	return ret, nil
}

func (p *Pipeline) AddFlow(flow Flow) (*FlowHandler, error) {
	ret := NewFlowHandler(flow)
	p.flows[len(p.flows)-1].Via(ret)
	p.flows = append(p.flows, ret)
	return ret, nil
}

func (p *Pipeline) Build() (returnErr error) {
	p.plumbingOnce.Do(func() {
		// inlets
		for kind, list := range p.Inlets {
			if reg := GetInletRegistry(kind); reg != nil {
				for _, v := range list {
					c := makeConfig(v, p.Defaults)
					flowCfgs := c.GetConfig("flows", nil)
					if flowCfgs != nil {
						c.Unset("flows")
					}
					inlet := reg.Factory(p.ctx.WithConfig(c))
					if _, err := p.AddInlet(kind, inlet); err != nil {
						p.ctx.LogError("failed to add inlet", "inlet", kind, "error", err.Error())
					} else {
						p.ctx.LogDebug(">>> flowsCfg", "type", fmt.Sprintf("%T", flowCfgs))
						for flowKind, flowCfg := range flowCfgs {
							p.ctx.LogDebug(">>> flow", "kind", flowKind, "cfg", fmt.Sprintf("%T", flowCfg))
						}
					}
				}
			} else {
				returnErr = fmt.Errorf("inlet %q not found", kind)
				return
			}
		}

		// flows
		for kind, list := range p.Flows {
			if reg := GetFlowRegistry(kind); reg != nil {
				for _, v := range list {
					c := makeConfig(v, p.Defaults)
					flow := reg.Factory(p.ctx.WithConfig(c))
					if _, err := p.AddFlow(flow); err != nil {
						p.ctx.LogError("failed to add flow", "flow", kind, "error", err.Error())
					}
				}
			} else {
				returnErr = fmt.Errorf("flow %q not found", kind)
				return
			}
		}

		// outlets
		for kind, list := range p.Outlets {
			if reg := GetOutletRegistry(kind); reg != nil {
				for _, v := range list {
					c := makeConfig(v, p.Defaults)
					outlet := reg.Factory(p.ctx.WithConfig(c))
					if _, err := p.AddOutlet(outlet); err != nil {
						p.ctx.LogError("failed to add outlet", "outlet", kind, "error", err.Error())
					}
				}
			} else {
				returnErr = fmt.Errorf("outlet %q not found", kind)
				return
			}
		}
	})
	return returnErr
}

// Start the pipeline, this will start all inlets, outlets and flows
// and returns immediately without waiting for the pipeline to stop
func (p *Pipeline) Start() {
	go p.Run()
}

// Run the pipeline, this will start all inlets, outlets and flows
// and wait until the pipeline is stopped
func (p *Pipeline) Run() error {
	// build pipeline
	if err := p.Build(); err != nil {
		p.ctx.LogError(p.logMsg("failed to build pipeline"), "error", err.Error())
		return err
	}
	// start outlets
	openOutlets := []*OutletHandler{}
	for _, out := range p.outputs {
		if err := out.Start(); err != nil {
			p.ctx.LogError("failed to start outlet", "error", err.Error())
		} else {
			openOutlets = append(openOutlets, out)
		}
	}
	p.outputs = openOutlets

	// fanout flow attached
	fanout := FanOutFlow(p.ctx).(*fanOutFlow)
	fanoutHandler := NewFlowHandler(fanout)
	fanout.LinkOutlets(p.outputs...)
	p.flows[len(p.flows)-1].Via(fanoutHandler)
	p.flows = append(p.flows, fanoutHandler)

	// start flows
	for _, flow := range p.flows {
		if err := flow.Start(); err != nil {
			// failed flow can not be allowed
			// it will break the pipeline
			p.ctx.LogError("failed to start flow", "error", err.Error())
			return err
		}
	}

	if len(p.inputs) == 0 {
		return errors.New("no inlet to start")
	}

	if len(p.outputs) == 0 {
		return errors.New("no outlet to start")
	}

	p.ctx.LogInfo("start", "inputs", len(p.inputs), "flows", len(p.flows)-2, "outputs", len(p.outputs))

	// start inputs
	inputWg := sync.WaitGroup{}
	for _, in := range p.inputs {
		inputWg.Add(1)
		go func(in *InletHandler) {
			defer inputWg.Done()
			if err := in.Run(); err != nil {
				p.ctx.LogError(p.logMsg("failed to start inlet"), "error", err.Error())
			}
		}(in)
	}

	inputWg.Wait()
	p.ctx.LogInfo("input completed")

	p.Stop()
	p.ctx.LogInfo(p.logMsg("stop"))

	return nil
}

// Stop the pipeline, this will stop all inlets, outlets and flows.
// Start() or Run() should be called before calling Stop()
func (p *Pipeline) Stop() error {
	p.stopOnce.Do(func() {
		for _, in := range p.inputs {
			in.Stop()
		}
		for _, fow := range p.flows {
			fow.Stop()
		}
		for _, out := range p.outputs {
			out.Stop()
		}
	})
	return nil
}

func makeConfig(v map[string]any, defaults map[string]any) Config {
	cfg := Config{}
	for k, v := range defaults {
		cfg[k] = v
	}
	for k, v := range v {
		cfg[k] = v
	}
	return cfg
}

//go:embed pipeline.toml
var DefaultConfigString string
