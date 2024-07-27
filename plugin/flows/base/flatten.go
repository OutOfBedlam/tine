package base

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

type flattenFlow struct {
	ctx       *engine.Context
	nameInfix string
}

func FlattenFlow(ctx *engine.Context) engine.Flow {
	nameInfix := ctx.Config().GetString("name_infix", "_")
	ret := &flattenFlow{ctx: ctx, nameInfix: nameInfix}
	return ret
}

func (ff *flattenFlow) Open() error      { return nil }
func (ff *flattenFlow) Close() error     { return nil }
func (ff *flattenFlow) Parallelism() int { return 1 }

func (ff *flattenFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	ret := []engine.Record{}
	for _, r := range recs {
		ts := r.Tags().Get(engine.TAG_TIMESTAMP)
		in := r.Tags().Get(engine.TAG_INLET)
		for _, f := range r.Fields() {
			fields := []*engine.Field{}
			if ts != nil && ts.IsNotNull() {
				fields = append(fields, engine.NewFieldWithValue(engine.TAG_TIMESTAMP, ts))
			}
			if in != nil && in.IsNotNull() {
				inStr, _ := in.String()
				fields = append(fields, engine.NewField("name", fmt.Sprintf("%s%s%s", inStr, ff.nameInfix, f.Name)))
			} else {
				fields = append(fields, engine.NewField("name", f.Name))
			}
			fields = append(fields, f.Copy("value"))
			ret = append(ret, engine.NewRecord(fields...))
		}
	}
	return ret, nil
}
