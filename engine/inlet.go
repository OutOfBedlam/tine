package engine

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
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

func InletWithPullFunc(fn func() ([]Record, error), interval time.Duration) Inlet {
	return &InletPullFuncWrap{fn: fn, interval: interval}
}

type InletPullFuncWrap struct {
	interval time.Duration
	fn       func() ([]Record, error)
}

var _ = PullInlet((*InletPullFuncWrap)(nil))

func (in *InletPullFuncWrap) Pull() ([]Record, error) {
	return in.fn()
}

func (in *InletPullFuncWrap) Interval() time.Duration {
	return 0
}

func (in *InletPullFuncWrap) Open() error {
	return nil
}

func (in *InletPullFuncWrap) Close() error {
	return nil
}

type InletHandler struct {
	name     string
	inlet    Inlet
	outCh    chan<- []Record
	interval *time.Ticker
	trigger  chan struct{}
	stopOnce sync.Once
	runner   func() error
	stopper  func()
}

func NewInletHandler(name string, inlet Inlet, outCh chan<- []Record) (*InletHandler, error) {
	if err := inlet.Open(); err != nil {
		return nil, fmt.Errorf("failed to open input: %v", err)
	}

	ret := &InletHandler{
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
	go in.runner()
}

func (in *InletHandler) Stop() {
	in.stopOnce.Do(func() {
		in.stopper()
	})
}

func (in *InletHandler) Run() error {
	return in.runner()
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

	for range in.trigger {
		if pull, ok := in.inlet.(PullInlet); ok {
			recs, err := pull.Pull()
			if len(recs) > 0 {
				in.outCh <- prependInletNameTimestamp(recs, in.name)
			}
			if err != nil {
				if err == io.EOF {
					slog.Debug("input eof")
				} else {
					slog.Error("failed to get input", "error", err.Error())
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
		slog.Error("failed to close input", "error", err.Error())
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
			slog.Error("inlethandler panic in inlet push", "error", e)
		}
	}()

	push.Push(func(recs []Record, err error) {
		if len(recs) > 0 {
			in.outCh <- prependInletNameTimestamp(recs, in.name)
		}
		if err != nil {
			if err == io.EOF {
				slog.Debug("input eof")
			} else {
				slog.Error("failed to get input", "error", err.Error())
			}
			in.Stop()
		}
	})
	if err := in.inlet.Close(); err != nil {
		slog.Error("failed to close input", "error", err.Error())
	}
	return nil
}

const FIELD_INLET = "_in"
const FIELD_TIMESTAMP = "_ts"

func prependInletNameTimestamp(recs []Record, name string) []Record {
	for i, r := range recs {
		recs[i] = NewRecord(
			NewTimeField(FIELD_TIMESTAMP, time.Now()),
			NewStringField(FIELD_INLET, name),
		).Append(r.Fields()...)
	}
	return recs
}
