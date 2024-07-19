package rrd

import (
	"fmt"
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
	path := ctx.Config().GetString("path", "")
	fields := ctx.Config().GetStringArray("fields", nil)
	timeField := ctx.Config().GetString("time_field", "_ts")
	return &rrdOutlet{
		ctx:       ctx,
		path:      path,
		fields:    fields,
		timeField: timeField,
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
	//startTime := o.ctx.Config().GetTime("start_time", time.Now())
	step := uint(o.ctx.Config().GetDuration("step", 10*time.Second).Seconds())
	heartbeat := 2 * step
	c := xrrd.NewCreator(o.path, time.Now().Add(-1*time.Second), step)
	for i, field := range o.fields {
		c.DS(fmt.Sprintf("%s=field%d[%d]", field, i, i), "GAUGE", heartbeat, 0, 10)
	}
	c.RRA("LAST", 0.5, 5, 100)
	c.RRA("AVERAGE", 0.5, 5, 100)
	if err := c.Create(true); err != nil {
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
			if field.Type == engine.TIME {
				values[i], _ = field.GetTime()
			} else if field.Type == engine.FLOAT {
				values[i], _ = field.GetFloat()
			} else if field.Type == engine.INT {
				values[i], _ = field.GetInt()
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
