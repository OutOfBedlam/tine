package expr

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
	goexpr "github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "expr",
		Factory: ExprFlow,
	})
}

func ExprFlow(ctx *engine.Context) engine.Flow {
	return &exprFlow{
		ctx: ctx,
	}
}

type exprFlow struct {
	ctx *engine.Context

	program *vm.Program
}

var _ = engine.Flow((*exprFlow)(nil))

func (ef *exprFlow) Open() error {
	code := ef.ctx.Config().GetString("program", "")
	if code == "" {
		return fmt.Errorf("program is empty")
	}

	if !doEval {
		prog, err := goexpr.Compile(code, goexpr.Env(DefaultProgramEnv))
		if err != nil {
			fmt.Println("compile error", err.Error())
			return err
		}
		// compiled program is safe for concurrent use.
		ef.program = prog
	}

	return nil
}

func (ef *exprFlow) Close() error {
	return nil
}

func (ef *exprFlow) Parallelism() int {
	return ef.ctx.Config().GetInt("parallelism", 1)
}

type ProgramEnv struct {
	Records []engine.Record
}

var DefaultProgramEnv = map[string]any{
	//"record":  map[string]any{},
	"records": []engine.Record{},
	"greet":   "Hello, %v",
	"sprintf": fmt.Sprintf,
	"println": fmt.Println,
}

const doEval = true

func (ef *exprFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	if doEval {
		env := DefaultProgramEnv
		env["records"] = recs
		code := ef.ctx.Config().GetString("program", "")
		output, err := goexpr.Eval(code, env)
		if err != nil {
			ef.ctx.LogError("flows.expr error\n" + err.Error())
			fmt.Println("expr error:", err)
			return nil, err
		}
		fmt.Println(">>", output)
	} else {
		env := DefaultProgramEnv
		env["records"] = recs
		output, err := goexpr.Run(ef.program, env)
		if err != nil {
			ef.ctx.LogError("flows.expr error\n" + err.Error())
			fmt.Println("expr error:", err)
			return nil, err
		}
		fmt.Println(">>", output)
	}

	return recs, nil
}
