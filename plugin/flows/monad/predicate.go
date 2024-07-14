package monad

import (
	"fmt"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
	goexpr "github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func ExprPredicate(code string) (engine.Predicate, error) {
	translated := []rune{}
	identBuff := []rune{}
	inIdent := false
	referedFiedls := []string{}

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
				referedFiedls = append(referedFiedls, ident)
				translated = append(translated, []rune(`_`+ident+".Value")...)
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
		referedFiedls:  referedFiedls,
	}

	// compile translated code
	env := map[string]any{}
	for _, rf := range referedFiedls {
		env["_"+rf] = (*engine.Field)(nil)
	}

	prog, err := goexpr.Compile(ret.translatedCode, goexpr.Env(env), goexpr.AsBool())
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
	referedFiedls  []string
	program        *vm.Program
	lastErr        error
}

func (ep *exprPredicate) Apply(record engine.Record) bool {
	env := map[string]any{}
	nonexists := []string{}
	for _, rf := range ep.referedFiedls {
		f := record.Field(rf)
		if f == nil {
			nonexists = append(nonexists, rf)
			continue
		}
		env["_"+rf] = f
	}
	if len(nonexists) > 0 {
		// TODO: If the record does not have the field, should we return false or error?
		//
		// For now, we return false. If we return error, the flow will be stopped.
		ep.lastErr = fmt.Errorf("fields not found: %v", nonexists)
		return false
	}

	result, err := goexpr.Run(ep.program, env)
	if err != nil {
		ep.lastErr = err
		return false
	}

	return result.(bool)
}
