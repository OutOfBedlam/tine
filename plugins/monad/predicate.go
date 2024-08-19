package monad

import (
	"fmt"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func ExprPredicate(code string) (engine.Predicate, error) {
	translated := []rune{}
	identBuff := []rune{}
	inIdent := false
	referredFields := []string{}
	referredVars := []string{}

	for i := 0; i < len(code); i++ {
		c := rune(code[i])
		switch c {
		case '$':
			if code[i+1] == '{' {
				inIdent = true
				i++
			}
		case '}':
			if inIdent {
				inIdent = false
				ident := strings.TrimSpace(string(identBuff))
				identBuff = identBuff[:0]
				referredFields = append(referredFields, ident)
				varName := "_" + strings.ReplaceAll(ident, ".", "_")
				referredVars = append(referredVars, varName)
				translated = append(translated, []rune(varName+".Value")...)
			} else {
				translated = append(translated, c)
			}
		default:
			if inIdent {
				identBuff = append(identBuff, c)
			} else {
				translated = append(translated, c)
			}
		}
	}

	ret := &exprPredicate{
		originalCode:   code,
		translatedCode: string(translated),
		referredFields: referredFields,
		referredVars:   referredVars,
	}

	// compile translated code
	env := map[string]any{}
	for _, rv := range referredVars {
		env[rv] = (*exprField)(nil)
	}

	prog, err := expr.Compile(ret.translatedCode, expr.Env(env), expr.AsBool())
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
		ep.lastErr = err
		return false
	}

	return result.(bool)
}
