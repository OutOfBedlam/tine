package base

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

type injectFlow struct {
	ctx    *engine.Context
	inject func([]engine.Record) ([]engine.Record, error)
}

func InjectFlow(ctx *engine.Context) engine.Flow {
	return &injectFlow{ctx: ctx}
}

func (f *injectFlow) Open() error {
	id := f.ctx.Config().GetString("id", "")
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if cb, ok := f.ctx.GetInject(id); ok {
		f.inject = cb
		return nil
	} else {
		return fmt.Errorf("flows.inject id %q not found", id)
	}
}

func (f *injectFlow) Close() error {
	return nil
}

func (f *injectFlow) Parallelism() int {
	return 1
}

func (f *injectFlow) Process(recs []engine.Record, nextFunc engine.FlowNextFunc) {
	if f.inject != nil {
		result, err := f.inject(recs)
		nextFunc(result, err)
	} else {
		nextFunc(recs, nil)
	}
}
