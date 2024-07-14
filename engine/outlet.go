package engine

import (
	"log/slog"
	"sync"
	"time"
)

type Outlet interface {
	OpenCloser
	Handle(r []Record) error
}

var outletRegistry map[string]*OutletReg = make(map[string]*OutletReg)
var outletNames = []string{}
var outletLock = sync.RWMutex{}

type OutletReg struct {
	Name    string
	Factory func(*Context) Outlet
}

func RegisterOutlet(reg *OutletReg) {
	outletLock.Lock()
	defer outletLock.Unlock()
	outletRegistry[reg.Name] = reg
	outletNames = append(outletNames, reg.Name)
}

func (or *OutletReg) New(ctx *Context) Outlet {
	return or.Factory(ctx)
}

func OutletNames() []string {
	outletLock.RLock()
	defer outletLock.RUnlock()
	return outletNames
}

func GetOutletRegistry(name string) *OutletReg {
	outletLock.RLock()
	defer outletLock.RUnlock()
	if reg, ok := outletRegistry[name]; ok {
		return reg
	}
	return nil
}

func OutletWithFunc(fn func([]Record) error) Outlet {
	return &OutletFuncWrap{fn: fn}
}

type OutletFuncWrap struct {
	fn func([]Record) error
}

func (out *OutletFuncWrap) Handle(r []Record) error {
	return out.fn(r)
}

func (out *OutletFuncWrap) Open() error {
	return nil
}

func (out *OutletFuncWrap) Close() error {
	return nil
}

type OutletHandler struct {
	inCh    chan []Record
	outlet  Outlet
	isOpen  bool
	closeCh chan bool
	closeWg sync.WaitGroup
	buffer  []Record
}

func NewOutletHandler(outlet Outlet) (*OutletHandler, error) {
	var interval time.Duration
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}
	ret := &OutletHandler{
		inCh:    make(chan []Record),
		outlet:  outlet,
		closeCh: make(chan bool),
	}
	return ret, nil
}

func (out *OutletHandler) Start() error {
	if err := out.outlet.Open(); err != nil {
		return err
	}

	out.closeWg.Add(1)
	go func() {
		out.isOpen = true
	loop:
		for {
			select {
			case r := <-out.inCh:
				out.buffer = append(out.buffer, r...)
				out.flush()
			case <-out.closeCh:
				out.flush()
				break loop
			}
		}
		out.closeWg.Done()
	}()
	return nil
}

func (out *OutletHandler) flush() {
	if len(out.buffer) == 0 {
		return
	}
	if err := out.outlet.Handle(out.buffer); err != nil {
		slog.Error("failed to output flush", "error", err.Error())
	}
	out.buffer = out.buffer[:0]
}

func (out *OutletHandler) Stop() {
	if !out.isOpen {
		return
	}
	out.closeCh <- true
	out.closeWg.Wait()
	close(out.inCh)

	if err := out.outlet.Close(); err != nil {
		slog.Error("failed to open output", "error", err.Error())
	}
}

func (out *OutletHandler) Sink() chan<- []Record {
	return out.inCh
}
