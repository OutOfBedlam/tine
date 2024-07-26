package base

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

type setFieldFlow struct {
	ctx        *engine.Context
	fields     []*engine.Field
	fieldNames []string
}

func SetFieldFlow(ctx *engine.Context) engine.Flow {
	ret := &setFieldFlow{ctx: ctx}
	return ret
}

func (sf *setFieldFlow) Open() error {
	set := sf.ctx.Config().GetConfigSlice("set", nil)
	if len(set) == 0 {
		return fmt.Errorf("set_field: no set field")
	}
	for _, s := range set {
		for key, val := range s {
			// TODO: the configuration only yields string values
			// if caller wants to replace _ts field via pipeline DSL, it is impossible for now
			switch v := val.(type) {
			case string:
				sf.fields = append(sf.fields, engine.NewStringField(key, v))
			case int:
				sf.fields = append(sf.fields, engine.NewIntField(key, int64(v)))
			case float64:
				sf.fields = append(sf.fields, engine.NewFloatField(key, v))
			case bool:
				sf.fields = append(sf.fields, engine.NewBoolField(key, v))
			default:
				sf.fields = append(sf.fields, engine.NewStringField(key, fmt.Sprintf("%v", v)))
			}
			sf.fieldNames = append(sf.fieldNames, key)
		}
	}
	return nil
}

func (sf *setFieldFlow) Close() error     { return nil }
func (sf *setFieldFlow) Parallelism() int { return 1 }

func (sf *setFieldFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	ret := []engine.Record{}
	for _, r := range recs {
		ret = append(ret, r.AppendOrReplace(sf.fields...))
	}
	return ret, nil
}
