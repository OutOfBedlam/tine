package engine

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Inlet interface {
	OpenCloser
	Process(InletNextFunc)
}

type InletNextFunc func([]Record, error)

type PeriodicInlet interface {
	Inlet
	Interval() time.Duration
}

var inletRegistry map[string]*InletReg = make(map[string]*InletReg)
var inletNames = []string{}
var inletLock = sync.RWMutex{}

type InletReg struct {
	Name    string
	Factory func(*Context) Inlet
}

func RegisterInlet(reg *InletReg) {
	inletLock.Lock()
	defer inletLock.Unlock()
	inletRegistry[reg.Name] = reg
	inletNames = append(inletNames, reg.Name)
}

func (ir *InletReg) New(ctx *Context) Inlet {
	return ir.Factory(ctx)
}

func InletNames() []string {
	inletLock.RLock()
	defer inletLock.RUnlock()
	return inletNames
}

func GetInletRegistry(name string) *InletReg {
	inletLock.RLock()
	defer inletLock.RUnlock()
	if reg, ok := inletRegistry[name]; ok {
		return reg
	}
	return nil
}

func InletWithFunc(fn func() ([]Record, error), opts ...InletFuncOption) Inlet {
	ret := &InletFuncWrap{fn: fn}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

type InletFuncWrap struct {
	interval      time.Duration
	runCountLimit int64
	runCount      int64
	fn            func() ([]Record, error)
}

type InletFuncOption func(*InletFuncWrap)

func WithInterval(interval time.Duration) InletFuncOption {
	return func(w *InletFuncWrap) {
		w.interval = interval
	}
}

func WithRunCountLimit(limit int64) InletFuncOption {
	return func(w *InletFuncWrap) {
		w.runCountLimit = limit
	}
}

var _ = Inlet((*InletFuncWrap)(nil))

func (in *InletFuncWrap) Process(next InletNextFunc) {
	count := atomic.AddInt64(&in.runCount, 1)
	if in.runCountLimit > 0 && count > in.runCountLimit {
		// already exceed the limit
		next(nil, io.EOF)
		return
	}
	recs, err := in.fn()
	if err != nil {
		next(recs, err)
		return
	}
	if in.runCountLimit > 0 && count >= in.runCountLimit {
		// it reaches the limit
		next(recs, io.EOF)
		return
	}
	next(recs, nil)
}

func (in *InletFuncWrap) Interval() time.Duration {
	return in.interval
}

func (in *InletFuncWrap) Open() error {
	return nil
}

func (in *InletFuncWrap) Close() error {
	return nil
}

type InletHandler struct {
	ctx      *Context
	name     string
	inlet    Inlet
	outCh    chan<- []Record
	interval *time.Ticker
	trigger  chan struct{}
	stopOnce sync.Once
	runner   func() error
	stopper  func()
	flows    []*FlowHandler
	sent     uint64
}

func NewInletHandler(ctx *Context, name string, inlet Inlet, outCh chan<- []Record) (*InletHandler, error) {
	if err := inlet.Open(); err != nil {
		return nil, fmt.Errorf("failed to open input: %v", err)
	}

	ret := &InletHandler{
		ctx:     ctx,
		name:    name,
		outCh:   outCh,
		inlet:   inlet,
		trigger: make(chan struct{}, 1),
	}

	interval := time.Duration(0)
	if ii, ok := inlet.(PeriodicInlet); ok {
		interval = ii.Interval()
	}
	if interval > 0 {
		ret.runner = ret.runPull
		ret.stopper = ret.stopPull
		if interval < 1*time.Second {
			interval = 1 * time.Second
		}
		ret.interval = time.NewTicker(interval)
	} else {
		ret.runner = ret.runPush
		ret.stopper = ret.stopPush
	}

	return ret, nil
}

func (in *InletHandler) Start() {
	in.startSubFlows()
	go in.runner()
}

func (in *InletHandler) Stop() {
	in.stopOnce.Do(func() {
		in.stopper()
		in.ctx.LogDebug("inlet stopped", "name", in.name, "sent", in.sent)
		in.stopSubFlows()
	})
}

func (in *InletHandler) Walk(walker func(inletName string, kind string, step string, handler any)) {
	for _, flow := range in.flows {
		walker(in.name, "flows", flow.name, flow)
	}
}

// AddFlow adds a sub-flow to the inlet handler
func (in *InletHandler) AddFlow(flow *FlowHandler) {
	in.flows = append(in.flows, flow)
}

// Via connects the inlet handler to a flow handler
func (in *InletHandler) Via(flow *FlowHandler) *FlowHandler {
	flow.outCh = in.outCh
	in.outCh = flow.inCh
	return flow
}

func (in *InletHandler) Run() error {
	in.startSubFlows()
	return in.runner()
}

func (in *InletHandler) startSubFlows() error {
	for _, f := range in.flows {
		if err := f.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (in *InletHandler) stopSubFlows() error {
	for _, f := range in.flows {
		f.Stop()
	}
	return nil
}

func (in *InletHandler) stopPull() {
	triggerOrg := in.trigger
	in.trigger = nil
	close(triggerOrg)
	if in.interval != nil {
		in.interval.Stop()
	}
}

func (in *InletHandler) runPull() error {
	defer func() {
		e := recover()
		if e != nil {
			if err, ok := e.(error); ok && !strings.Contains(err.Error(), "send on closed channel") {
				in.ctx.LogError("panic recover", "error", err)
			} else {
				in.ctx.LogWarn("panic recover", "error", e)
			}
		}
	}()

	go func() {
		in.trigger <- struct{}{}
		for range in.interval.C {
			if in.trigger == nil {
				break
			}
			in.trigger <- struct{}{}
		}
	}()

	for range in.trigger {
		doBreak := false
		in.inlet.Process(func(recs []Record, err error) {
			if len(recs) > 0 {
				in.outCh <- prependInletNameTimestamp(recs, in.name)
				atomic.AddUint64(&in.sent, uint64(len(recs)))
			}
			if err != nil {
				if err == io.EOF {
					in.ctx.LogDebug("input eof")
				} else {
					in.ctx.LogError("failed to get input", "error", err.Error())
				}
				doBreak = true
			}
		})
		if doBreak {
			break
		}
	}
	in.Stop()
	return nil
}

func (in *InletHandler) stopPush() {
	if err := in.inlet.Close(); err != nil {
		in.ctx.LogError("failed to close input", "error", err.Error())
	}
}

func (in *InletHandler) runPush() error {
	defer func() {
		e := recover()
		if e != nil {
			in.ctx.LogError("inlethandler panic in inlet push", "error", e)
		}
	}()

	in.inlet.Process(func(recs []Record, err error) {
		if len(recs) > 0 {
			in.outCh <- prependInletNameTimestamp(recs, in.name)
			atomic.AddUint64(&in.sent, uint64(len(recs)))
		}
		if err != nil {
			if err == io.EOF {
				in.ctx.LogDebug("input eof")
			} else {
				in.ctx.LogError("failed to get input", "error", err.Error())
			}
			in.Stop()
		}
	})
	if err := in.inlet.Close(); err != nil {
		in.ctx.LogError("failed to close input", "error", err.Error())
	}
	return nil
}

const TAG_INLET = "_in"
const TAG_TIMESTAMP = "_ts"

func prependInletNameTimestamp(recs []Record, name string) []Record {
	for _, r := range recs {
		r.Tags().Set(TAG_INLET, NewValue(name))
		r.Tags().Set(TAG_TIMESTAMP, NewValue(Now()))
	}
	return recs
}

// Now is a function that returns the current time
// It is used to generate the timestamp for each record
// The default value is time.Now
// Set this to a fixed time in the purpose of testing
var Now = time.Now
