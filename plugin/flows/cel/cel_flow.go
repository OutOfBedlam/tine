package cel

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	gocel "github.com/google/cel-go/cel"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{Name: "cel", Factory: NewCelFlow})
}

func NewCelFlow(ctx *engine.Context) engine.Flow {
	return &CelFlow{
		ctx: ctx,
	}
}

type CelFlow struct {
	ctx     *engine.Context
	program gocel.Program
}

func (cf *CelFlow) Open() error {
	code := cf.ctx.Config().GetString("program", "")
	if code == "" {
		return fmt.Errorf("program is empty")
	}
	env, err := gocel.NewEnv(
		gocel.Variable("now", gocel.TimestampType),
	)
	if err != nil {
		return err
	}

	ast, iss := env.Compile(code)
	if iss != nil && iss.Err() != nil {
		return err
	}

	checked, iss := env.Check(ast)
	if iss != nil && iss.Err() != nil {
		return err
	}

	prg, err := env.Program(checked)
	if err != nil {
		return err
	}

	cf.program = prg
	return nil
}

func (cf *CelFlow) Close() error {
	return nil
}

func (cf *CelFlow) Parallelism() int {
	return cf.ctx.Config().GetInt("parallelism", 1)
}

func (cf *CelFlow) Process(records []engine.Record) ([]engine.Record, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("CEL panic", "error", fmt.Sprintf("%v", r))
		}
	}()
	out, _, err := cf.program.Eval(map[string]any{
		"now": time.Now(),
	})
	if err != nil {
		slog.Error("CEL", "error", err.Error())
		return nil, err
	}
	records = append(records, engine.NewRecord(
		engine.NewField("NAME", "CEL"),
		engine.NewField("VALUE", fmt.Sprintf("%v", out.Value())),
	))
	return records, nil
}
