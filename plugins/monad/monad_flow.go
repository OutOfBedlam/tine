package monad

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "filter",
		Factory: FilterFlow,
	})
}

func FilterFlow(ctx *engine.Context) engine.Flow {
	return &filterFlow{
		ctx: ctx,
	}
}

type filterFlow struct {
	ctx       *engine.Context
	predicate engine.Predicate
}

var _ = engine.Flow((*filterFlow)(nil))

func (ff *filterFlow) Open() error {
	strPredicate := ff.ctx.Config().GetString("predicate", "")
	if strPredicate == "" {
		return fmt.Errorf("predicate is empty")
	}
	if pred, err := ExprPredicate(strPredicate); err != nil {
		return err
	} else {
		ff.predicate = pred
	}
	return nil
}

func (ff *filterFlow) Close() error {
	return nil
}

func (ff *filterFlow) Process(records []engine.Record, nextFunc engine.FlowNextFunc) {
	if ff.predicate == nil {
		nextFunc(records, nil)
		return
	}
	ret := make([]engine.Record, 0, len(records))
	for _, record := range records {
		result := ff.predicate.Apply(record)
		if result {
			ret = append(ret, record)
		}
	}
	nextFunc(ret, nil)
}

func (ff *filterFlow) Parallelism() int {
	return 1
}
