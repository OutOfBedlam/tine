package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

type dumpFlow struct {
	ctx    *engine.Context
	vf     engine.ValueFormat
	logger func(string, ...any)
}

func DumpFlow(ctx *engine.Context) engine.Flow {
	return &dumpFlow{ctx: ctx}
}

func (df *dumpFlow) Open() error {
	conf := df.ctx.Config()
	level := conf.GetString("level", "DEBUG")
	switch strings.ToUpper(level) {
	case "DEBUG":
		df.logger = df.ctx.LogDebug
	case "INFO":
		df.logger = df.ctx.LogInfo
	case "WARN":
		df.logger = df.ctx.LogWarn
	case "ERROR":
		df.logger = df.ctx.LogError
	default:
		df.logger = df.ctx.LogDebug
	}
	decimal := conf.GetInt("decimal", -1)
	timeformat := conf.GetString("timeformat", "2006-01-02 15:04:05")
	tz := conf.GetString("tz", "Local")
	if z, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("flow-dump invalid timezone %q, %w", tz, err)
	} else {
		df.vf = engine.ValueFormat{
			Timeformat: engine.NewTimeformatterWithLocation(timeformat, z),
			Decimal:    decimal,
		}
	}
	return nil
}

func (df *dumpFlow) Close() error     { return nil }
func (df *dumpFlow) Parallelism() int { return 1 }

func (df *dumpFlow) Process(recs []engine.Record, nextFunc engine.FlowNextFunc) {
	for idx, r := range recs {
		list := make([]any, 0, len(r.Fields())*2)
		for _, f := range r.Fields() {
			list = append(list, f.Name)
			list = append(list, f.Value.Format(df.vf))
		}
		df.logger("flow-dump", append([]any{"rec", fmt.Sprintf("%d/%d", idx+1, len(recs))}, list...)...)
	}
	nextFunc(recs, nil)
}
