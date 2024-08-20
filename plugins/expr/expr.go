package expr

import (
	"fmt"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "filter",
		Factory: FilterFlow,
	})
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "map",
		Factory: MapFlow,
	})
}

type Translated struct {
	OriginalCode   string
	Code           string
	ReferredFields []string
	ReferredVars   []string
}

func Translate(code string) *Translated {
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

	return &Translated{
		OriginalCode:   code,
		Code:           string(translated),
		ReferredFields: referredFields,
		ReferredVars:   referredVars,
	}
}

func (trans *Translated) makeEnv() map[string]any {
	env := map[string]any{
		"printf":  fmt.Printf,
		"println": fmt.Println,
	}
	return env
}

func (trans *Translated) EmptyEnv() map[string]any {
	env := trans.makeEnv()
	for _, rv := range trans.ReferredVars {
		env[rv] = (*exprField)(nil)
	}
	return env
}

func (trans *Translated) RecordEnv(record engine.Record) (map[string]any, error) {
	env := trans.makeEnv()
	nonExists := []string{}
	for idx, rf := range trans.ReferredFields {
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
		varName := trans.ReferredVars[idx]
		env[varName] = ef
	}
	if len(nonExists) > 0 {
		// TODO: If the record does not have the field, should we return false or error?
		//
		// For now, we return false. If we return error, the pipeline will be stopped.
		return nil, fmt.Errorf("fields not found: %v", nonExists)
	}
	return env, nil
}
