package engine

import (
	"sync"
	"sync/atomic"
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
	ctx     *Context
	name    string
	inCh    chan []Record
	outlet  Outlet
	isOpen  bool
	closeCh chan bool
	closeWg sync.WaitGroup
	buffer  []Record
	recvCnt uint64
	doneCnt uint64
}

func NewOutletHandler(ctx *Context, name string, outlet Outlet) (*OutletHandler, error) {
	var interval time.Duration
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}
	ret := &OutletHandler{
		ctx:     ctx,
		name:    name,
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
		out.ctx.LogDebug("outlet started", "name", out.name)
	loop:
		for {
			select {
			case r := <-out.inCh:
				out.buffer = append(out.buffer, r...)
				atomic.AddUint64(&out.recvCnt, uint64(len(r)))
				out.flush(false)
			case <-out.closeCh:
				break loop
			}
		}
		out.flush(true)
		out.isOpen = false
		out.closeWg.Done()
	}()
	return nil
}

func (out *OutletHandler) flush(_ /*force*/ bool) {
	if len(out.buffer) == 0 {
		return
	}
	if err := out.outlet.Handle(out.buffer); err != nil {
		out.ctx.LogError("failed to output flush", "error", err.Error())
	} else {
		atomic.AddUint64(&out.doneCnt, uint64(len(out.buffer)))
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
		out.ctx.LogError("failed to open output", "error", err.Error())
	}
	out.ctx.LogDebug("outlet stopped", "name", out.name, "recv", out.recvCnt, "done", out.doneCnt)
}

func (out *OutletHandler) Sink() chan<- []Record {
	return out.inCh
}
