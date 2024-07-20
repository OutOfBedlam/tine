package engine

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
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
	rawWriter    io.Writer

	setContentTypeFunc     SetContentTypeCallback
	setContentEncodingFunc SetContentEncodingCallback
}

type SetContentTypeCallback func(contentType string)
type SetContentEncodingCallback func(contentEncoding string)

type Option func(*Pipeline) error

var pipelineSerial int64

func WithConfig(conf string) Option {
	return func(p *Pipeline) error {
		if err := LoadConfig(conf, &p.PipelineConfig); err != nil {
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
			if err := LoadConfig(string(content), &p.PipelineConfig); err != nil {
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

func WithWriter(w io.Writer) Option {
	return func(p *Pipeline) error {
		p.rawWriter = w
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(p *Pipeline) error {
		p.logger = logger
		return nil
	}
}

func WithSetContentTypeFunc(fn SetContentTypeCallback) Option {
	return func(p *Pipeline) error {
		p.setContentTypeFunc = fn
		return nil
	}
}

func WithSetContentEncodingFunc(fn SetContentEncodingCallback) Option {
	return func(p *Pipeline) error {
		p.setContentEncodingFunc = fn
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
	p.flows = []*FlowHandler{NewFlowHandler(p.ctx, "fan-in", FanInFlow(p.ctx))}
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
	ret, err := NewInletHandler(p.ctx, name, inlet, p.flows[0].inCh)
	if err != nil {
		return nil, err
	}
	p.inputs = append(p.inputs, ret)
	return ret, nil
}

func (p *Pipeline) AddOutlet(name string, outlet Outlet) (*OutletHandler, error) {
	ret, err := NewOutletHandler(p.ctx, name, outlet)
	if err != nil {
		return nil, err
	}
	p.outputs = append(p.outputs, ret)
	return ret, nil
}

func (p *Pipeline) AddFlow(name string, flow Flow) (*FlowHandler, error) {
	ret := NewFlowHandler(p.ctx, name, flow)
	p.flows[len(p.flows)-1].Via(ret)
	p.flows = append(p.flows, ret)
	return ret, nil
}

func (p *Pipeline) Build() (returnErr error) {
	p.plumbingOnce.Do(func() {
		// inlets
		for _, inletCfg := range p.Inlets {
			if reg := GetInletRegistry(inletCfg.Plugin); reg != nil {
				c := makeConfig(inletCfg.Params, p.Defaults)
				inlet := reg.Factory(p.ctx.WithConfig(c))
				if inletHandler, err := p.AddInlet(inletCfg.Plugin, inlet); err != nil {
					p.ctx.LogError("failed to add inlet", "inlet", inletCfg.Plugin, "error", err.Error())
				} else {
					var lastFlow *FlowHandler
					for _, flowCfg := range inletCfg.Flows {
						reg := GetFlowRegistry(flowCfg.Plugin)
						if reg == nil {
							returnErr = fmt.Errorf("flow %q not found", flowCfg.Plugin)
							return
						}
						c := makeConfig(flowCfg.Params, p.Defaults)
						flow := NewFlowHandler(p.ctx, flowCfg.Plugin, reg.Factory(p.ctx.WithConfig(c)))
						if lastFlow == nil {
							lastFlow = inletHandler.Via(flow)
						} else {
							lastFlow = lastFlow.Via(flow)
						}
						inletHandler.AddFlow(flow)
					}
				}
			} else {
				returnErr = fmt.Errorf("inlet %q not found", inletCfg.Plugin)
				return
			}
		}

		// flows
		for _, flowCfg := range p.Flows {
			if reg := GetFlowRegistry(flowCfg.Plugin); reg != nil {
				c := makeConfig(flowCfg.Params, p.Defaults)
				flow := reg.Factory(p.ctx.WithConfig(c))
				if _, err := p.AddFlow(flowCfg.Plugin, flow); err != nil {
					p.ctx.LogError("failed to add flow", "flow", flowCfg.Plugin, "error", err.Error())
				}
			} else {
				returnErr = fmt.Errorf("flow %q not found", flowCfg.Plugin)
				return
			}
		}

		// outlets
		for _, outletCfg := range p.Outlets {
			if reg := GetOutletRegistry(outletCfg.Plugin); reg != nil {
				c := makeConfig(outletCfg.Params, p.Defaults)
				outlet := reg.Factory(p.ctx.WithConfig(c))
				if _, err := p.AddOutlet(outletCfg.Plugin, outlet); err != nil {
					p.ctx.LogError("failed to add outlet", "outlet", outletCfg.Plugin, "error", err.Error())
				}
			} else {
				returnErr = fmt.Errorf("outlet %q not found", outletCfg.Plugin)
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
	fanoutHandler := NewFlowHandler(p.ctx, "fan-out", fanout)
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

	p.ctx.LogInfo("start", "inlets", len(p.inputs), "flows", len(p.flows)-2, "outlets", len(p.outputs))

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
	p.ctx.LogDebug("inlets completed")

	p.Stop()
	p.ctx.LogInfo("stop")

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
