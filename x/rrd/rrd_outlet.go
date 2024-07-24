package rrd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	xrrd "github.com/OutOfBedlam/tine/x/rrd/internal/rrd"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "rrd",
		Factory: RRDOutlet,
	})
}

func RRDOutlet(ctx *engine.Context) engine.Outlet {
	return &rrdOutlet{
		ctx: ctx,
	}
}

type rrdOutlet struct {
	ctx       *engine.Context
	path      string
	fields    []string
	timeField string
	updater   *xrrd.Updater
}

func (o *rrdOutlet) Open() error {
	o.path = o.ctx.Config().GetString("path", "")
	o.timeField = o.ctx.Config().GetString("time_field", "_ts")
	step := uint(o.ctx.Config().GetDuration("step", 10*time.Second).Seconds())
	if step < 1 {
		step = 1
	}
	defaultHeartbeat := o.ctx.Config().GetDuration("heartbeat", time.Duration(2*step)*time.Second)
	overwrite := o.ctx.Config().GetBool("overwrite", false)

	c := xrrd.NewCreator(o.path, time.Now().Add(-1*time.Second), step)

	fieldsCfg := o.ctx.Config().GetConfigArray("fields", nil)
	for _, fc := range fieldsCfg {
		// field
		field := fc.GetString("field", "")
		o.fields = append(o.fields, field)
		// ds
		ds := fc.GetString("ds", field)
		ds = strings.ReplaceAll(ds, ":", "_")
		// dst  GAUGE, COUNTER, DCOUNTER, DERIVE, DDERIVE, ABSOLUTE, COMPUTE
		dst := fc.GetString("dst", "GAUGE")
		// heartbeat
		heartbeat := fc.GetDuration("heartbeat", defaultHeartbeat)
		// min, max
		min := fc.GetString("min", "U")
		max := fc.GetString("max", "U")
		// rpn
		if rpn := fc.GetString("rpn", ""); rpn != "" {
			c.DS(ds, dst, heartbeat.Seconds(), min, max, rpn)
		} else {
			c.DS(ds, dst, heartbeat.Seconds(), min, max)
		}
	}

	rralst := o.ctx.Config().GetConfigArray("rra", nil)
	for _, rra := range rralst {
		var args = []any{}
		// CF
		cf := strings.ToUpper(rra.GetString("cf", ""))
		if cf != "AVERAGE" && cf != "MIN" && cf != "MAX" && cf != "LAST" {
			return fmt.Errorf("invalid rra cf=%q", cf)
		}
		// xff
		xff := rra.GetFloat("xff", 0.5)
		args = append(args, xff)
		// steps
		steps := rra.GetString("steps", "1")
		args = append(args, steps)
		// rows
		rows := rra.GetString("rows", "1000")
		args = append(args, rows)
		c.RRA(cf, args...)
	}

	if err := c.Create(overwrite); err != nil && !os.IsExist(err) {
		return fmt.Errorf("fail to create rrd path=%q, %s", o.path, err.Error())
	}
	o.updater = xrrd.NewUpdater(o.path)
	return nil
}

func (o *rrdOutlet) Close() error {
	if inf, err := xrrd.Info(o.path); err == nil {
		o.ctx.LogDebug("outlets.rrd", "info", inf)
	}

	return nil
}

func (o *rrdOutlet) Handle(recs []engine.Record) error {
	fieldNames := append([]string{o.timeField}, o.fields...)
	for _, rec := range recs {
		fields := rec.Fields(fieldNames...)

		if len(fields) != len(fieldNames) || fields[0] == nil {
			continue
		}
		values := make([]any, len(fields))
		for i, field := range fields {
			if field == nil {
				o.ctx.LogWarn("outlets.rrd update, field not found", "field", fieldNames[i])
				goto next_record
			}
			if field.Type() == engine.TIME {
				values[i], _ = field.Value.Time()
			} else if field.Type() == engine.FLOAT {
				values[i], _ = field.Value.Float64()
			} else if field.Type() == engine.INT {
				values[i], _ = field.Value.Int64()
			} else {
				o.ctx.LogWarn("outlets.rrd update, unsupported field type", "field", fieldNames[i])
				return nil
			}
		}
		o.updater.Cache(values...)
	next_record:
	}
	err := o.updater.Update()
	if err != nil {
		o.ctx.LogWarn("outlets.rrd update", "error", err)
	}
	return nil
}
