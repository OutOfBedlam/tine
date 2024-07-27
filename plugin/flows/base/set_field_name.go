package base

import (
	"github.com/OutOfBedlam/tine/engine"
)

func SetFieldNameFlow(ctx *engine.Context) engine.Flow {
	prefix := ctx.Config().GetString("prefix", "")
	suffix := ctx.Config().GetString("suffix", "")
	ret := &setFieldNameFlow{
		prefix:   prefix,
		suffix:   suffix,
		replaces: make(map[string]string),
	}

	replaces := ctx.Config()["replaces"]
	if replaces != nil {
		if replaces, ok := replaces.([]any); ok {
			for _, repl := range replaces {
				if item, ok := repl.([]any); ok && len(item) == 2 {
					oldName := item[0].(string)
					newName := item[1].(string)
					ret.replaces[oldName] = newName
				}
			}
		}
	}
	return ret
}

type setFieldNameFlow struct {
	prefix   string
	suffix   string
	replaces map[string]string
}

var _ = engine.Flow((*setFieldNameFlow)(nil))

func (nf *setFieldNameFlow) Open() error      { return nil }
func (nf *setFieldNameFlow) Close() error     { return nil }
func (nf *setFieldNameFlow) Parallelism() int { return 1 }

func (nf *setFieldNameFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	if len(nf.replaces) > 0 {
		for _, r := range recs {
			for _, f := range r.Fields() {
				if newName, ok := nf.replaces[f.Name]; ok {
					f.Name = newName
				}
			}
		}
	}
	if nf.prefix != "" || nf.suffix != "" {
		for _, r := range recs {
			for _, f := range r.Fields() {
				f.Name = nf.prefix + f.Name + nf.suffix
			}
		}
	}
	return recs, nil
}
