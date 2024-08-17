package engine

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/OutOfBedlam/tine/util"
)

type OpenCloser interface {
	Open() error
	Close() error
}

type Pipeline struct {
	PipelineConfig
	inputs     []*InletHandler
	outputs    []*OutletHandler
	flows      []*FlowHandler
	ctx        *Context
	logger     *slog.Logger
	logWriter  io.Writer
	logVerbose bool
	buildOnce  sync.Once
	startOnce  sync.Once
	stopOnce   sync.Once
	rawWriter  io.Writer

	setContentTypeFunc     SetContentTypeCallback
	setContentEncodingFunc SetContentEncodingCallback
	setContentLengthFunc   SetContentLengthCallback
}

type SetContentTypeCallback func(contentType string)
type SetContentEncodingCallback func(contentEncoding string)
type SetContentLengthCallback func(contentLength int)

type Option func(*Pipeline) error

var pipelineSerial int64

// WithConfig loads a TOML configuration string into a PipelineConfig struct
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

// WithConfigFile loads a TOML configuration file into a PipelineConfig struct
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

// WithConfigTemplate loads a TOML configuration template into a PipelineConfig struct
func WithConfigTemplate(configTemplate string, data map[string][]string) Option {
	return func(p *Pipeline) error {
		params := map[string]any{}
		for k, v := range data {
			if len(v) == 0 {
				params[k] = ""
			} else if len(v) == 1 {
				params[k] = v[0]
			} else {
				params[k] = v
			}
		}
		tmpl, err := template.New("pipeline").Parse(configTemplate)
		if err != nil {
			return err
		}
		buff := &bytes.Buffer{}
		if err := tmpl.Execute(buff, params); err != nil {
			return err
		}
		if err := LoadConfig(buff.String(), &p.PipelineConfig); err != nil {
			return err
		}
		if p.Name == "" {
			serial := atomic.AddInt64(&pipelineSerial, 1)
			p.Name = fmt.Sprintf("pipeline-%d", serial)
		}
		return nil
	}
}

// WithName sets the name of the pipeline
func WithName(name string) Option {
	return func(p *Pipeline) error {
		p.Name = name
		return nil
	}
}

// WithDefaults sets the default output writer for the pipeline
func WithWriter(w io.Writer) Option {
	return func(p *Pipeline) error {
		if rspWriter, ok := w.(http.ResponseWriter); ok {
			p.setHttpResponseWriter(rspWriter)
		} else {
			p.rawWriter = w
		}
		return nil
	}
}

// WithLogger sets the logger for the pipeline
func WithLogger(logger *slog.Logger) Option {
	return func(p *Pipeline) error {
		p.logger = logger
		return nil
	}
}

func WithLogWriter(w io.Writer) Option {
	return func(p *Pipeline) error {
		p.logWriter = w
		return nil
	}
}

func WithVerbose(flag bool) Option {
	return func(p *Pipeline) error {
		p.logVerbose = flag
		return nil
	}
}

// WithSetContentTypeFunc sets the callback function to set the content type
func WithSetContentTypeFunc(fn SetContentTypeCallback) Option {
	return func(p *Pipeline) error {
		p.setContentTypeFunc = fn
		return nil
	}
}

// WithSetContentEncodingFunc sets the callback function to set the content encoding
func WithSetContentEncodingFunc(fn SetContentEncodingCallback) Option {
	return func(p *Pipeline) error {
		p.setContentEncodingFunc = fn
		return nil
	}
}

// WithSetContentLengthFunc sets the callback function to set the content length
func WithSetContentLengthFunc(fn SetContentLengthCallback) Option {
	return func(p *Pipeline) error {
		p.setContentLengthFunc = fn
		return nil
	}
}

// New creates a new pipeline with the given options
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
	if p.logVerbose {
		p.Log.Level = "DEBUG"
		if p.Log.Path == "" {
			p.Log.Path = "-"
		}
	}

	// init logging
	if p.logger == nil {
		if p.logWriter != nil {
			p.logger = util.NewLoggerWithWriter(p.Log, p.logWriter)
		} else {
			p.logger = util.NewLogger(p.Log)
		}
	}

	// context
	p.ctx = newContext(p).WithLogger(p.logger)
	// add fan-in flow
	// If the pipeline is built by programmatically, the fan-in flow should exist.
	// TODO: Add fan-in flow only when it is needed.
	p.flows = []*FlowHandler{NewFlowHandler(p.ctx, "fan-in", FanInFlow(p.ctx))}
	return p, nil
}

// HttpHandleFunc is a convenience function to create a http.HandlerFunc
// from a pipeline configuration
func HttpHandleFunc(config string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := New(WithConfig(config), WithWriter(w))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = p.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.Stop()
	}
}

func (p *Pipeline) setHttpResponseWriter(w http.ResponseWriter) {
	p.rawWriter = w
	p.setContentTypeFunc = func(contentType string) {
		w.Header().Set("Content-Type", contentType)
	}
	p.setContentEncodingFunc = func(contentEncoding string) {
		w.Header().Set("Content-Encoding", contentEncoding)
	}
	p.setContentLengthFunc = func(contentLength int) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	}
}

func (p *Pipeline) logMsg(msg string) string {
	if p == nil {
		return fmt.Sprintf("pipeline %s", msg)
	} else {
		return fmt.Sprintf("pipeline %s %s", p.Name, msg)
	}
}

// Context returns the context of the pipeline
func (p *Pipeline) Context() *Context {
	return p.ctx
}

// AddInlet adds an inlet to the pipeline
func (p *Pipeline) AddInlet(name string, inlet Inlet) (*InletHandler, error) {
	ret, err := NewInletHandler(p.ctx, name, inlet, p.flows[0].inCh)
	if err != nil {
		return nil, err
	}
	p.inputs = append(p.inputs, ret)
	return ret, nil
}

// AddOutlet adds an outlet to the pipeline
func (p *Pipeline) AddOutlet(name string, outlet Outlet) (*OutletHandler, error) {
	ret, err := NewOutletHandler(p.ctx, name, outlet)
	if err != nil {
		return nil, err
	}
	p.outputs = append(p.outputs, ret)
	return ret, nil
}

// AddFlow adds a flow to the pipeline
func (p *Pipeline) AddFlow(name string, flow Flow) (*FlowHandler, error) {
	ret := NewFlowHandler(p.ctx, name, flow)
	p.flows[len(p.flows)-1].Via(ret)
	p.flows = append(p.flows, ret)
	return ret, nil
}

// Build the pipeline, this will create all inlets, outlets and flows
func (p *Pipeline) Build() (returnErr error) {
	p.buildOnce.Do(func() {
		// inlets
		for _, inletCfg := range p.Inlets {
			if reg := GetInletRegistry(inletCfg.Plugin); reg != nil {
				c := makeConfig(inletCfg.Params, p.Defaults)
				inlet := reg.Factory(p.ctx.WithConfig(c))
				if inletHandler, err := p.AddInlet(inletCfg.Plugin, inlet); err != nil {
					p.ctx.LogError("failed to add inlet", "inlet", inletCfg.Plugin, "error", err.Error())
					returnErr = err
					return
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

		// fan-out flow attached
		// TODO: Add fan-out flow only when it is needed.
		fanOut := FanOutFlow(p.ctx).(*fanOutFlow)
		fanOutHandler := NewFlowHandler(p.ctx, "fan-out", fanOut)
		p.flows[len(p.flows)-1].Via(fanOutHandler)
		p.flows = append(p.flows, fanOutHandler)

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
	return
}

func (p *Pipeline) Walk(walker func(pipelineName string, kind string, step string, handler any)) {
	for _, input := range p.inputs {
		walker(p.Name, "inlets", input.name, input)
	}
	for _, flow := range p.flows {
		walker(p.Name, "flows", flow.name, flow)
	}
	for _, output := range p.outputs {
		walker(p.Name, "outlets", output.name, output)
	}
}

// Start the pipeline, this will start all inlets, outlets and flows
// and returns immediately without waiting for the pipeline to stop
func (p *Pipeline) Start() {
	go p.Run()
}

// Run the pipeline, this will start all inlets, outlets and flows
// and wait until the pipeline is stopped
func (p *Pipeline) Run() (returnErr error) {
	p.startOnce.Do(func() {
		returnErr = p.run0()
	})
	return returnErr
}

// Do not call this method directly, use Run() and Start() instead
func (p *Pipeline) run0() error {
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

	// fanOut flow attached
	if fanOut, ok := p.flows[len(p.flows)-1].flow.(*fanOutFlow); ok {
		fanOut.LinkOutlets(p.outputs...)
	}

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

	p.ctx.LogInfo("start", "inlets", len(p.inputs), "flows", len(p.flows), "outlets", len(p.outputs))

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
