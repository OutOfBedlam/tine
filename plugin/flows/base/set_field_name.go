package base

import (
	"github.com/OutOfBedlam/tine/engine"
)

func SetFieldNameFlow(ctx *engine.Context) engine.Flow {
	prefix := ctx.Config().GetString("prefix", "")
	suffix := ctx.Config().GetString("prefix", "")
	return &setFieldNameFlow{
		prefix: prefix,
		suffix: suffix,
	}
}

type setFieldNameFlow struct {
	prefix string
	suffix string
}

var _ = engine.Flow((*setFieldNameFlow)(nil))

func (nf *setFieldNameFlow) Open() error      { return nil }
func (nf *setFieldNameFlow) Close() error     { return nil }
func (nf *setFieldNameFlow) Parallelism() int { return 1 }

func (nf *setFieldNameFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	if nf.prefix == "" && nf.suffix == "" {
		return recs, nil
	}

	for _, r := range recs {
		for _, f := range r.Fields() {
			f.Name = nf.prefix + f.Name + nf.suffix
		}
	}
	return recs, nil
}
