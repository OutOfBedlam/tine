package engine

import (
	"sync"
	"sync/atomic"
)

var flowRegistry = make(map[string]*FlowReg)
var flowsLock sync.RWMutex

type Flow interface {
	OpenCloser
	Process([]Record) ([]Record, error)
	Parallelism() int
}

type FlowReg struct {
	Name    string
	Factory func(ctx *Context) Flow
}

func RegisterFlow(reg *FlowReg) {
	flowsLock.Lock()
	defer flowsLock.Unlock()
	flowRegistry[reg.Name] = reg
}

func UnregisterFlow(name string) {
	flowsLock.Lock()
	defer flowsLock.Unlock()
	delete(flowRegistry, name)
}

func GetFlowRegistry(name string) *FlowReg {
	flowsLock.RLock()
	defer flowsLock.RUnlock()
	if f, ok := flowRegistry[name]; ok {
		return f
	}
	return nil
}

func FlowNames() []string {
	flowsLock.RLock()
	defer flowsLock.RUnlock()
	ret := make([]string, 0, len(flowRegistry))
	for k := range flowRegistry {
		ret = append(ret, k)
	}
	return ret
}

type FlowHandler struct {
	ctx   *Context
	name  string
	inCh  chan []Record
	outCh chan<- []Record
	flow  Flow

	parallelism chan struct{}
	closeWg     sync.WaitGroup

	recv uint64
	sent uint64
}

func NewFlowHandler(ctx *Context, name string, flow Flow) *FlowHandler {
	parallelism := flow.Parallelism()
	if parallelism < 1 {
		parallelism = 1
	}
	ret := &FlowHandler{
		ctx:  ctx,
		name: name,
		inCh: make(chan []Record),
		flow: flow,
	}
	if parallelism > 1 {
		ret.parallelism = make(chan struct{}, parallelism)
	}
	return ret
}

func (fh *FlowHandler) Via(next *FlowHandler) *FlowHandler {
	next.outCh = fh.outCh
	fh.outCh = next.inCh
	return next
}

func (fh *FlowHandler) Start() error {
	if err := fh.flow.Open(); err != nil {
		return err
	}
	fh.closeWg.Add(1)

	if fh.parallelism != nil {
		go func() {
			for records := range fh.inCh {
				atomic.AddUint64(&fh.recv, uint64(len(records)))
				fh.closeWg.Add(1)
				fh.parallelism <- struct{}{}
				go func(r []Record) {
					defer func() {
						fh.closeWg.Done()
						<-fh.parallelism
					}()
					r, err := fh.flow.Process(r)
					if err != nil {
						fh.ctx.LogError("failed to handle flow", "error", err.Error())
					}
					if len(r) > 0 {
						fh.outCh <- r
						atomic.AddUint64(&fh.sent, uint64(len(r)))
					}
				}(records)
			}
			fh.closeWg.Done()
		}()
	} else {
		go func() {
			for records := range fh.inCh {
				atomic.AddUint64(&fh.recv, uint64(len(records)))
				r, err := fh.flow.Process(records)
				if err != nil {
					fh.ctx.LogError("failed to handle flow", "error", err.Error())
				}
				if len(r) > 0 {
					fh.outCh <- r
					atomic.AddUint64(&fh.sent, uint64(len(r)))
				} else if _, ok := fh.flow.(*fanOutFlow); ok {
					atomic.AddUint64(&fh.sent, uint64(len(records)))
				}
			}
			fh.closeWg.Done()
		}()
	}
	return nil
}

func (fh *FlowHandler) Stop() error {
	close(fh.inCh)
	fh.closeWg.Wait()
	if fh.parallelism != nil {
		close(fh.parallelism)
	}
	if err := fh.flow.Close(); err != nil {
		return err
	}
	fh.ctx.LogDebug("flow stopped", "name", fh.name, "recv", fh.recv, "sent", fh.sent)
	return nil
}

func init() {
	RegisterFlow(&FlowReg{Name: "fan-in", Factory: FanInFlow})
	RegisterFlow(&FlowReg{Name: "fan-out", Factory: FanOutFlow})
}

type fanInFlow struct {
}

var _ = (Flow)((*fanInFlow)(nil))

func FanInFlow(ctx *Context) Flow {
	return &fanInFlow{}
}

func (ff *fanInFlow) Open() error      { return nil }
func (ff *fanInFlow) Close() error     { return nil }
func (ff *fanInFlow) Parallelism() int { return 1 }

func (ff *fanInFlow) Process(r []Record) ([]Record, error) { return r, nil }

type fanOutFlow struct {
	outs []chan<- []Record
}

func FanOutFlow(ctx *Context) Flow {
	return &fanOutFlow{}
}

func (ff *fanOutFlow) Open() error      { return nil }
func (ff *fanOutFlow) Close() error     { return nil }
func (ff *fanOutFlow) Parallelism() int { return 1 }

func (ff *fanOutFlow) LinkOutlets(outs ...*OutletHandler) {
	for _, o := range outs {
		ff.outs = append(ff.outs, o.inCh)
	}
}

func (ff *fanOutFlow) Process(r []Record) ([]Record, error) {
	for _, o := range ff.outs {
		o <- r
	}
	return nil, nil
}

type FlowFuncWrapOption func(*FlowFuncWrap)

func WithFlowFuncParallelism(parallelism int) FlowFuncWrapOption {
	return func(fw *FlowFuncWrap) {
		if parallelism > 0 {
			fw.parallelism = 1
		} else {
			fw.parallelism = parallelism
		}
	}
}

type FlowFuncWrap struct {
	fn          func([]Record) ([]Record, error)
	parallelism int
}

var _ = (Flow)((*FlowFuncWrap)(nil))

func FlowWithFunc(fn func([]Record) ([]Record, error), opts ...FlowFuncWrapOption) Flow {
	ret := &FlowFuncWrap{fn: fn}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (fw *FlowFuncWrap) Open() error {
	return nil
}

func (fw *FlowFuncWrap) Close() error {
	return nil
}

func (fw *FlowFuncWrap) Process(r []Record) ([]Record, error) {
	return fw.fn(r)
}

func (fw *FlowFuncWrap) Parallelism() int {
	return fw.parallelism
}
