package exec

import (
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "exec",
		Factory: ExecInlet,
	})
}

func ExecInlet(ctx *engine.Context) engine.Inlet {
	return &execInlet{ExecDriver: New(ctx)}
}

type execInlet struct {
	*ExecDriver
}

func (ei *execInlet) Process(nextFunc engine.InletNextFunc) {
	ei.ExecDriver.Process(nil, nextFunc)
}
