package base

import (
	"slices"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

type selectFlow struct {
	ctx      *engine.Context
	includes []string
}

func SelectFlow(ctx *engine.Context) engine.Flow {
	includes := ctx.Config().GetStringSlice("includes", []string{"#*", "*"})
	for i := 0; i < len(includes); i++ {
		if includes[i] == "**" {
			includes[i] = "#*"
			i++
			includes = slices.Insert(includes, i, "*")
		}
	}
	return &selectFlow{ctx: ctx, includes: includes}
}

func (sf *selectFlow) Open() error      { return nil }
func (sf *selectFlow) Close() error     { return nil }
func (sf *selectFlow) Parallelism() int { return 1 }

func (sf *selectFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	ret := []engine.Record{}
	for _, r := range recs {
		fields := []*engine.Field{}
		for _, item := range sf.includes {
			if item == "*" {
				fields = append(fields, r.Fields()...)
			} else if strings.HasPrefix(item, "#") {
				tag := item[1:]
				if tag == "*" {
					names := r.Tags().Names()
					slices.Sort(names)
					for _, nm := range names {
						if v := r.Tags().Get(nm); v != nil {
							fields = append(fields, engine.NewFieldWithValue(nm, v))
						} else {
							fields = append(fields, engine.NewFieldWithValue(nm, engine.NewUntypedNullValue()))
						}
					}
				} else {
					if v := r.Tags().Get(tag); v != nil {
						fields = append(fields, engine.NewFieldWithValue(tag, v))
					} else {
						fields = append(fields, engine.NewFieldWithValue(tag, engine.NewUntypedNullValue()))
					}
				}
			} else {
				f := r.Field(item)
				if f == nil {
					fields = append(fields, engine.NewFieldWithValue(item, engine.NewUntypedNullValue()))
				} else {
					fields = append(fields, f)
				}
			}
		}
		rec := engine.NewRecord(fields...)
		rec.Tags().Merge(r.Tags())
		ret = append(ret, rec)
	}
	return ret, nil
}
