package expr

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func ExprPredicate(code string) (engine.Predicate, error) {
	trans := Translate(code)
	ret := &exprPredicate{
		originalCode:   trans.OriginalCode,
		translatedCode: trans.Code,
		referredFields: trans.ReferredFields,
		referredVars:   trans.ReferredVars,
	}

	// compile translated code
	prog, err := expr.Compile(ret.translatedCode, expr.Env(trans.EmptyEnv()), expr.AsBool())
	if err != nil {
		return ret, err
	} else {
		ret.program = prog
	}
	return ret, nil
}

type exprPredicate struct {
	originalCode   string
	translatedCode string
	referredFields []string
	referredVars   []string
	program        *vm.Program
	lastErr        error
}

type exprField struct {
	Name   string
	Type   string
	IsNull bool
	Value  any
}

func (ep *exprPredicate) Apply(record engine.Record) bool {
	env := map[string]any{}
	nonExists := []string{}
	for idx, rf := range ep.referredFields {
		f := record.Field(rf)
		if f == nil {
			nonExists = append(nonExists, rf)
			continue
		}
		ef := &exprField{
			Name:   f.Name,
			Type:   f.Type().String(),
			IsNull: f.IsNull(),
			Value:  f.Value.Raw(),
		}
		varName := ep.referredVars[idx]
		env[varName] = ef
	}
	if len(nonExists) > 0 {
		// TODO: If the record does not have the field, should we return false or error?
		//
		// For now, we return false. If we return error, the pipeline will be stopped.
		ep.lastErr = fmt.Errorf("fields not found: %v", nonExists)
		return false
	}

	result, err := expr.Run(ep.program, env)
	if err != nil {
		fmt.Println("--->", err)
		ep.lastErr = err
		return false
	}

	return result.(bool)
}
