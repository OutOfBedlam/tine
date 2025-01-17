package rrd

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"
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
	interval := ctx.Config().GetDuration("interval", 0)
	if interval <= 0 {
		interval = 0
	} else if interval < time.Second {
		interval = time.Second
	}

	return &rrdGraphInlet{
		ctx:      ctx,
		interval: interval,
	}
}

type rrdGraphInlet struct {
	ctx      *engine.Context
	interval time.Duration
	runLimit int64
	runCount int64
}

var _ = (engine.Inlet)((*rrdGraphInlet)(nil))

func (ri *rrdGraphInlet) Open() error             { return nil }
func (ri *rrdGraphInlet) Close() error            { return nil }
func (ri *rrdGraphInlet) Interval() time.Duration { return ri.interval }
func (ri *rrdGraphInlet) Process(next engine.InletNextFunc) {
	runCount := atomic.AddInt64(&ri.runCount, 1)
	if ri.runLimit > 0 && runCount > ri.runLimit {
		next(nil, io.EOF)
		return
	}
	recs, err := generate(ri.ctx)
	if ri.runLimit > 0 && runCount >= ri.runLimit {
		err = io.EOF
	}
	next(recs, err)
}

func generate(ctx *engine.Context) ([]engine.Record, error) {
	conf := ctx.Config()
	defaultPath := conf.GetString("path", "")

	g := xrrd.NewGrapher()
	if title := conf.GetString("title", ""); title != "" {
		g.SetTitle(title)
	}
	if vLabel := conf.GetString("v_label", ""); vLabel != "" {
		g.SetVLabel(vLabel)
	}
	if size := conf.GetIntSlice("size", []int{800, 300}); len(size) == 2 && size[0]*size[1] > 0 {
		g.SetSize(uint(size[0]), uint(size[1]))
	} else {
		ctx.LogWarn("inlets.rrd_graph", "invalid size", size)
	}
	if wm := conf.GetString("watermark", ctx.PipelineName()); wm != "" {
		g.SetWatermark(wm)
	}
	if ul := conf.GetInt("units_length", 5); ul > 0 {
		g.SetUnitsLength(uint(ul))
	}
	if border := uint(conf.GetInt("border", 0)); border >= uint(0) {
		g.SetBorder(border)
	}
	if ue := int(conf.GetInt("units_exponent", math.MinInt64)); ue != math.MinInt64 {
		g.SetUnitsExponent(ue)
	}
	options := []string{"--disable-rrdtool-tag"}
	if zoom := conf.GetFloat("zoom", 0); zoom > 0 {
		options = append(options, "--zoom", fmt.Sprintf("%f", zoom))
	}
	if len(options) > 0 {
		g.AddOptions(options...)
	}

	var theme *Theme
	if th, ok := themes[conf.GetString("theme", "")]; ok {
		theme = th
	} else {
		theme = themes["grayscale"]
	}

	g.SetColor("BACK", strings.TrimPrefix(theme.Back, "#"))
	g.SetColor("CANVAS", strings.TrimPrefix(theme.Canvas, "#"))
	g.SetColor("FONT", strings.TrimPrefix(theme.Font, "#"))

	seriesNameMaxLen := 0
	series := conf.GetConfigSlice("fields", nil)
	for _, ser := range series {
		ds := ser.GetString("ds", "")
		vname := ser.GetString("vname", ds)
		if len(vname) > seriesNameMaxLen {
			seriesNameMaxLen = len(vname)
		}
	}

	for i, ser := range series {
		path := ser.GetString("path", defaultPath)
		ds := ser.GetString("ds", "")
		vname := ser.GetString("vname", ds)
		typ := strings.ToUpper(ser.GetString("type", "LINE"))
		color := ser.GetString("color", theme.Palettes[i%len(theme.Palettes)])
		color = strings.TrimPrefix(color, "#")
		if len(color) == 6 {
			if typ == "LINE" {
				color += "ff"
			} else {
				color += "66"
			}
		}
		if len(color) != 8 {
			return nil, fmt.Errorf("invalid color: %s", color)
		}
		cf := ser.GetString("cf", "AVERAGE")

		g.Def(vname, path, ds, cf)

		nameFormat := fmt.Sprintf("%%-%ds", seriesNameMaxLen)
		nameFormat = ser.GetString("name", nameFormat)
		nameFormatted := fmt.Sprintf(nameFormat, vname)
		switch typ {
		case "LINE":
			width := float32(ser.GetFloat("width", 1.0))
			g.Line(width, vname, color, nameFormatted)
		case "AREA":
			g.Area(vname, color, nameFormatted)
		}

		gprintOrder := []string{"min", "max", "avg", "last"}
		gprintFn := []string{"MINIMUM", "MAXIMUM", "AVERAGE", "LAST"}
		for i, gp := range gprintOrder {
			fn := gprintFn[i]
			if f := ser.GetString(gp, ""); f != "" {
				g.VDef(vname+"_"+gp, vname+","+fn) // g.VDef("load_cur", "load,LAST")
				g.GPrint(vname+"_"+gp, gp+" "+f)   // g.GPrint("load_cur", "last %4.2lf\\n")
			}
		}
	}

	now := time.Now()
	dur := conf.GetDuration("range", 600*time.Second)
	nfo, buf, err := g.Graph(now.Add(dur*-1), now)
	if err != nil {
		return nil, err
	}

	ctx.LogDebug("inlets.rrd_graph", "info", fmt.Sprintf("%+v", nfo))
	bv := engine.NewField("graph", buf)
	bv.Tags.Set(engine.CanonicalTagKey("Content-Type"), engine.NewValue("image/png"))

	recs := []engine.Record{
		engine.NewRecord(bv),
	}
	return recs, nil
}

type Theme struct {
	Back     string
	Canvas   string
	Font     string
	Palettes []string
}

var themes = map[string]*Theme{
	"nonamed": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#003f5c", "#2f4b7c", "#665191", "#a05195", "#d45087", "#f95d6a", "#ff7c43", "#ffa600", "#f0f921", "#b3d334"},
	},
	"retro_metro": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#ea5545", "#f46a9b", "#ef9b20", "#edbf33", "#ede15b", "#bdcf32", "#87bc45", "#27aeef", "#b33dc6", "#f15a29"},
	},
	"dutch_field": {
		Back:     "#000000",
		Canvas:   "#000000",
		Font:     "#ffffff",
		Palettes: []string{"#e60049", "#0bb4ff", "#50e991", "#e6d800", "#9b19f5", "#ffa300", "#dc0ab4", "#b3d4ff", "#00bfa0", "#ff6e00"},
	},
	"river_nights": {
		Back:     "#000000",
		Canvas:   "#000000",
		Font:     "#ffffff",
		Palettes: []string{"#b30000", "#7c1158", "#4421af", "#1a53ff", "#0d88e6", "#00b7c7", "#5ad45a", "#8be04e", "#ebdc78", "#ffcc00"},
	},
	"spring_pastels": {
		Back:     "#000000",
		Canvas:   "#000000",
		Font:     "#ffffff",
		Palettes: []string{"#fd7f6f", "#7eb0d5", "#b2e061", "#bd7ebe", "#ffb55a", "#ffee65", "#beb9db", "#fdcce5", "#8bd3c7", "#ff9e9d"},
	},
	"grayscale": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#0d0d0d", "#262626", "#404040", "#595959", "#737373", "#8c8c8c", "#a6a6a6", "#bfbfbf", "#d9d9d9", "#f2f2f2"},
	},
	"qualitative": {
		Back:     "#000000",
		Canvas:   "#000000",
		Font:     "#ffffff",
		Palettes: []string{"#cecece", "#a559aa", "#59a89c", "#f0c571", "#e02b35", "#082a54"},
	},
	"bright": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#003a7d", "#008dff", "#ff73b6", "#c701ff", "#4ecb8d", "#ff9d3a", "#f9e858", "#d83034"},
	},
	"muted": {
		Back:     "#000000",
		Canvas:   "#000000",
		Font:     "#ffffff",
		Palettes: []string{"#f0c571", "#59a89c", "#0b81a2", "#e25759", "#9d2c00", "#7e4794", "#36b700"},
	},
	"wctu": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#EA644A", "#EC9D48", "#ECD748", "#54EC48", "#48C4EC", "#68E4FC", "#7648EC", "#DE48EC"},
	},
	"wctu2": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#CC3118", "#CC7016", "#C9B215", "#24BC14", "#1598C3", "#0588B3", "#4D18E4", "#B415C7"},
	},
	"gchart": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#3366CC", "#FF9900", "#990099", "#0099C6", "#66AA00", "#316395", "#22AA99", "#6633CC", "#8B0707", "#5574A6"},
	},
	"gchart2": {
		Back:     "#ffffff",
		Canvas:   "#ffffff",
		Font:     "#000000",
		Palettes: []string{"#DC3912", "#109618", "#3B3EAC", "#DD4477", "#B82E2E", "#994499", "#AAAA11", "#E67300", "#329262", "#651067"},
	},
}
