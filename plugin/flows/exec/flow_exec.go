package exec

import (
	"github.com/OutOfBedlam/tine/drivers/exec"
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "exec",
		Factory: ExecFlow,
	})
}

func ExecFlow(ctx *engine.Context) engine.Flow {
	return &execFlow{ExecDriver: exec.New(ctx)}
}

type execFlow struct {
	*exec.ExecDriver
}

var _ = engine.Flow((*execFlow)(nil))

func (ef *execFlow) Process(records []engine.Record, nextFunc engine.FlowNextFunc) {
	for _, rec := range records {
		ef.ExecDriver.Process(rec, nextFunc)
	}
}
