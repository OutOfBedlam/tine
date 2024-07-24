package base

import (
	"fmt"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

type dumpFlow struct {
	ctx       *engine.Context
	precision int
	logger    func(string, ...any)
	tf        *engine.Timeformatter
}

func DumpFlow(ctx *engine.Context) engine.Flow {
	ret := &dumpFlow{ctx: ctx}
	ret.precision = ctx.Config().GetInt("precision", -1)
	level := ctx.Config().GetString("level", "DEBUG")
	switch strings.ToUpper(level) {
	case "DEBUG":
		ret.logger = ctx.LogDebug
	case "INFO":
		ret.logger = ctx.LogInfo
	case "WARN":
		ret.logger = ctx.LogWarn
	case "ERROR":
		ret.logger = ctx.LogError
	default:
		ret.logger = ctx.LogDebug
	}
	ret.tf = engine.NewTimeformatter(ctx.Config().GetString("timeformat", "2006-01-02 15:04:05"))
	return ret
}

func (df *dumpFlow) Open() error      { return nil }
func (df *dumpFlow) Close() error     { return nil }
func (df *dumpFlow) Parallelism() int { return 1 }

func (df *dumpFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	formatOpt := engine.FormatOption{Timeformat: df.tf, Decimal: df.precision}
	for idx, r := range recs {
		list := make([]any, 0, len(r.Fields())*2)
		for _, f := range r.Fields() {
			list = append(list, f.Name)
			list = append(list, f.Value.Format(formatOpt))
		}
		df.logger("flow-dump", append([]any{"rec", fmt.Sprintf("%d/%d", idx+1, len(recs))}, list...)...)
	}
	return recs, nil
}
