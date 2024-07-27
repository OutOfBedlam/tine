package engine

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Inlet OpenCloser

type PullInlet interface {
	Inlet
	Pull() ([]Record, error)
	Interval() time.Duration
}

type PushInlet interface {
	Inlet
	Push(func([]Record, error))
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

func InletWithPullFunc(fn func() ([]Record, error), opts ...InletPullFuncOption) Inlet {
	ret := &InletPullFuncWrap{fn: fn}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

type InletPullFuncWrap struct {
	interval      time.Duration
	runCountLimit int64
	runCount      int64
	fn            func() ([]Record, error)
}

type InletPullFuncOption func(*InletPullFuncWrap)

func WithInterval(interval time.Duration) InletPullFuncOption {
	return func(w *InletPullFuncWrap) {
		w.interval = interval
	}
}

func WithRunCountLimit(limit int64) InletPullFuncOption {
	return func(w *InletPullFuncWrap) {
		w.runCountLimit = limit
	}
}

var _ = PullInlet((*InletPullFuncWrap)(nil))

func (in *InletPullFuncWrap) Pull() ([]Record, error) {
	count := atomic.AddInt64(&in.runCount, 1)
	if in.runCountLimit > 0 && count > in.runCountLimit {
		// already exceed the limit
		return nil, io.EOF
	}
	recs, err := in.fn()
	if err != nil {
		return recs, err
	}
	if in.runCountLimit > 0 && count >= in.runCountLimit {
		// it reaches the limit
		return recs, io.EOF
	}
	return recs, nil
}

func (in *InletPullFuncWrap) Interval() time.Duration {
	return in.interval
}

func (in *InletPullFuncWrap) Open() error {
	return nil
}

func (in *InletPullFuncWrap) Close() error {
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

	if _, ok := inlet.(PushInlet); ok {
		ret.runner = ret.runPush
		ret.stopper = ret.stopPush
	} else if pullInlet, ok := inlet.(PullInlet); ok {
		ret.runner = ret.runPull
		ret.stopper = ret.stopPull
		interval := pullInlet.Interval()
		if interval < 1*time.Second {
			interval = 1 * time.Second
		}
		ret.interval = time.NewTicker(interval)
	} else {
		return nil, fmt.Errorf("unsupported inlet type: %T", inlet)
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
	go func() {
		in.trigger <- struct{}{}
		for range in.interval.C {
			if in.trigger == nil {
				break
			}
			in.trigger <- struct{}{}
		}
	}()

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
	for range in.trigger {
		if pull, ok := in.inlet.(PullInlet); ok {
			recs, err := pull.Pull()
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
				break
			}
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
	push, ok := in.inlet.(PushInlet)
	if !ok {
		return fmt.Errorf("inlet does not support push")
	}

	defer func() {
		e := recover()
		if e != nil {
			in.ctx.LogError("inlethandler panic in inlet push", "error", e)
		}
	}()

	push.Push(func(recs []Record, err error) {
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
