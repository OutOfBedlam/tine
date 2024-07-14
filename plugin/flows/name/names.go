package name

import (
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{Name: "name-prefix", Factory: NamePrefixFlow})
}

func NamePrefixFlow(ctx *engine.Context) engine.Flow {
	prefix := ctx.Config().GetString("prefix", "")
	parallelism := ctx.Config().GetInt("parallelism", 1)
	return &namePrefixFlow{
		prefix:      prefix,
		parallelism: parallelism,
	}
}

type namePrefixFlow struct {
	prefix      string
	parallelism int
}

var _ = engine.Flow((*namePrefixFlow)(nil))

func (nf *namePrefixFlow) Open() error      { return nil }
func (nf *namePrefixFlow) Close() error     { return nil }
func (cf *namePrefixFlow) Parallelism() int { return cf.parallelism }

func (nf *namePrefixFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	if nf.prefix == "" {
		return recs, nil
	}

	for _, r := range recs {
		for _, f := range r.Fields() {
			if f.Name == "NAME" && f.Type == engine.STRING {
				f.Value = nf.prefix + f.Value.(string)
			}
		}
	}
	return recs, nil
}
