package expr

import (
	"fmt"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func MapFlow(ctx *engine.Context) engine.Flow {
	return &mapFlow{
		ctx: ctx,
	}
}

type mapFlow struct {
	ctx         *engine.Context
	predicate   engine.Predicate
	code        string
	leftValue   string
	program     *vm.Program
	translated  *Translated
	parallelism int
}

var _ = engine.Flow((*mapFlow)(nil))

func (mf *mapFlow) Open() error {
	mf.parallelism = mf.ctx.Config().GetInt("parallelism", 1)

	strPredicate := mf.ctx.Config().GetString("predicate", "")
	if strPredicate != "" {
		if pred, err := ExprPredicate(strPredicate); err != nil {
			return err
		} else {
			mf.predicate = pred
		}
	}

	mf.code = mf.ctx.Config().GetString("code", "")
	if mf.code == "" {
		return fmt.Errorf("code is empty")
	}

	code := ""
	if idx := strings.IndexRune(mf.code, '='); idx > 0 {
		mf.leftValue = strings.TrimSpace(mf.code[0:idx])
		code = strings.TrimSpace(mf.code[idx+1:])
	} else {
		code = mf.code
	}
	mf.translated = Translate(code)
	prog, err := expr.Compile(mf.translated.Code, expr.Env(mf.translated.EmptyEnv()))
	if err != nil {
		return err
	} else {
		mf.program = prog
	}
	return nil
}

func (mf *mapFlow) Close() error {
	return nil
}

func (mf *mapFlow) Process(records []engine.Record, nextFunc engine.FlowNextFunc) {
	if mf.code == "" {
		nextFunc(records, nil)
		return
	}
	for i, record := range records {
		env, err := mf.translated.RecordEnv(record)
		if err != nil {
			nextFunc(nil, err)
			return
		}
		if mf.predicate != nil {
			if !mf.predicate.Apply(record) {
				continue
			}
		}
		result, err := expr.Run(mf.program, env)
		if err != nil {
			nextFunc(nil, err)
			return
		}
		if mf.leftValue != "" {
			fv := record.Field(mf.leftValue)
			if fv == nil {
				nextFunc(nil, fmt.Errorf("left value field %s not found", mf.leftValue))
				return
			}
			switch v := result.(type) {
			case int:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, int64(v)))
			case int64:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			case float64:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			case string:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			case bool:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			case time.Time:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			case []byte:
				records[i] = record.AppendOrReplace(engine.NewField(mf.leftValue, v))
			default:
				nextFunc(nil, fmt.Errorf("expression result is unsupported type %T", v))
			}
		}
	}
	nextFunc(records, nil)
}

func (mf *mapFlow) Parallelism() int {
	return mf.parallelism
}
