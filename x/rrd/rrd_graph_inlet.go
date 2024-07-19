package rrd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	xrrd "github.com/OutOfBedlam/tine/x/rrd/internal/rrd"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "rrd_graph",
		Factory: RRDGraphInlet,
	})
}

func RRDGraphInlet(ctx *engine.Context) engine.Inlet {
	return &rrdGraphInlet{
		ctx: ctx,
	}
}

type rrdGraphInlet struct {
	ctx *engine.Context
}

var _ = (engine.PushInlet)((*rrdGraphInlet)(nil))

func (ri *rrdGraphInlet) Open() error {
	return nil
}

func (ri *rrdGraphInlet) Close() error {
	return nil
}

func (ri *rrdGraphInlet) Push(cb func([]engine.Record, error)) {
	conf := ri.ctx.Config()
	defaultPath := conf.GetString("path", "")

	g := xrrd.NewGrapher()
	if title := conf.GetString("title", ""); title != "" {
		g.SetTitle(title)
	}
	if vLabel := conf.GetString("v_label", ""); vLabel != "" {
		g.SetVLabel(vLabel)
	}
	if size := conf.GetIntArray("size", []int{800, 300}); len(size) == 2 && size[0]*size[1] > 0 {
		g.SetSize(uint(size[0]), uint(size[1]))
	} else {
		ri.ctx.LogWarn("inlets.rrd_graph", "invalid size", size)
	}
	if wm := conf.GetString("watermark", ri.ctx.PipelineName()); wm != "" {
		g.SetWatermark(wm)
	}
	if ul := conf.GetInt("units_length", 5); ul > 0 {
		g.SetUnitsLength(uint(ul))
	}
	series := conf.GetConfigArray("series", nil)
	for _, ser := range series {
		path := ser.GetString("path", defaultPath)
		ds := ser.GetString("ds", "")
		vname := ser.GetString("vname", ds)
		typ := ser.GetString("type", "LINE")
		color := ser.GetString("color", "0000FF")
		cf := ser.GetString("cf", "AVERAGE")

		// g.Def("v1", path, "load1", "AVERAGE")
		// g.Def("v2", path, "load5", "AVERAGE")
		// g.VDef("max1", "v1,MAXIMUM")
		// g.VDef("avg2", "v2,AVERAGE")
		// g.GPrintT("max1", "max1 at %c")
		// g.GPrint("avg2", "avg2=%lf")
		// g.PrintT("max1", "max1 at %c")
		// g.Print("avg2", "avg2=%lf")
		g.Def(vname, path, ds, cf)
		switch strings.ToUpper(typ) {
		case "LINE":
			g.Line(1, vname, color, vname)
		case "AREA":
			g.Area(vname, color, vname)
		}
	}

	now := time.Now()
	dur := conf.GetDuration("range", 600*time.Second)
	nfo, buf, err := g.Graph(now.Add(dur*-1), now)
	if err != nil {
		cb(nil, err)
	}

	ri.ctx.LogDebug("inlets.rrd_graph", "info", fmt.Sprintf("%+v", nfo))
	bv := engine.NewBinaryValue(buf)
	bv.SetHeader("Content-Type", "image/png")

	recs := []engine.Record{
		engine.NewRecord(
			engine.NewBinaryField("graph", bv),
		),
	}
	cb(recs, io.EOF)
}
