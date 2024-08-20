package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/OutOfBedlam/tine/util"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "template",
		Factory: TemplateOutlet,
	})
}

func TemplateOutlet(ctx *engine.Context) engine.Outlet {
	ret := &templateOutlet{ctx: ctx}
	return ret
}

type templateOutlet struct {
	ctx          *engine.Context
	tmpl         *template.Template
	writer       io.WriteCloser
	valueFormat  engine.ValueFormat
	columnSeries string
	lazy         bool
	rowNum       int64
	table        *engine.Table[int64]
}

func (to *templateOutlet) Open() error {
	conf := to.ctx.Config()
	to.columnSeries = strings.ToLower(conf.GetString("column_series", ""))
	to.lazy = conf.GetBool("lazy", false)
	to.valueFormat = engine.DefaultValueFormat()
	if tf := conf.GetString("timeformat", ""); tf != "" {
		tz, err := time.LoadLocation(conf.GetString("tz", "Local"))
		if err != nil {
			return err
		}
		to.valueFormat.Timeformat = engine.NewTimeformatterWithLocation(tf, tz)
	} else {
		tz, err := time.LoadLocation(conf.GetString("tz", "Local"))
		if err != nil {
			return err
		}
		to.valueFormat.Timeformat = engine.NewTimeformatterWithLocation(time.RFC3339, tz)
	}
	to.valueFormat.Decimal = conf.GetInt("decimal", -1)

	templates := conf.GetStringSlice("templates", []string{})
	templateFiles := conf.GetStringSlice("template_files", []string{})

	if len(templates) == 0 && len(templateFiles) == 0 {
		return fmt.Errorf("no template provided")
	}
	to.tmpl = template.New("template").Funcs(template.FuncMap{
		"timeformat": func(t time.Time, layout string) string {
			switch layout {
			case "ns":
				return fmt.Sprintf("%d", t.UnixNano())
			case "us":
				return fmt.Sprintf("%d", t.UnixNano()/1000)
			case "ms":
				return fmt.Sprintf("%d", t.UnixNano()/1000000)
			case "s":
				return fmt.Sprintf("%d", t.Unix())
			default:
				return t.Format(layout)
			}
		},
	})

	var err error
	for _, t := range templates {
		to.tmpl, err = to.tmpl.Parse(t)
		if err != nil {
			return err
		}
	}
	if len(templateFiles) > 0 {
		to.tmpl, err = to.tmpl.ParseFiles(templateFiles...)
		if err != nil {
			return err
		}
	}

	contentType := conf.GetString("content_type", "text/plain")
	to.ctx.SetContentType(contentType)

	if w := to.ctx.Writer(); w != nil {
		to.writer = engine.NopCloser(w)
	} else {
		path := conf.GetString("path", "")
		switch path {
		case "":
			to.writer = engine.NopCloser(io.Discard)
		case "-":
			to.writer = engine.NopCloser(os.Stdout)
		default:
			overwrite := conf.GetBool("overwrite", false)
			if _, err := os.Stat(path); err == nil && !overwrite {
				return os.ErrExist
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			to.writer = f
		}
	}
	return nil
}

func (to *templateOutlet) Close() error {
	if to.columnSeries != "" && to.table != nil {
		to.closeTable()
	}
	if to.writer != nil {
		to.writer.Close()
	}
	return nil
}

func (to *templateOutlet) Handle(recs []engine.Record) error {
	if to.columnSeries != "" {
		return to.handleTable(recs)
	} else {
		return to.handleRecords(recs)
	}
}

func (to *templateOutlet) handleRecords(recs []engine.Record) error {
	arr := []map[string]any{}
	for _, rec := range recs {
		fields := rec.Fields()
		obj := map[string]any{}
		for k, v := range rec.Tags() {
			obj[k] = v.Raw()
		}
		for _, f := range fields {
			obj[f.Name] = f.Value.Raw()
		}
		arr = append(arr, obj)
	}
	err := to.tmpl.Execute(to.writer, arr)
	if err != nil {
		return err
	}
	return nil
}

func (to *templateOutlet) handleTable(recs []engine.Record) error {
	if to.table == nil {
		to.table = engine.NewTable[int64]()
	}
	for _, rec := range recs {
		to.rowNum++
		to.table.Set(to.rowNum, rec.Fields()...)
	}
	if !to.lazy {
		to.closeTable()
		to.table = nil
	}
	return nil
}

func (to *templateOutlet) closeTable() error {
	if to.table == nil {
		return nil
	}
	cols := to.table.Columns()
	obj := map[string]any{}
	for _, col := range cols {
		arr := []any{}
		for _, v := range to.table.Series(col) {
			arr = append(arr, v.Value.Raw())
		}
		if to.columnSeries == "json" {
			js, _ := json.Marshal(to.seriesJson(arr))
			obj[col] = string(js)
		} else {
			obj[col] = arr
		}
	}
	err := to.tmpl.Execute(to.writer, obj)
	if err != nil {
		return err
	}
	return nil
}

func (to *templateOutlet) seriesJson(val any) any {
	switch v := val.(type) {
	case []any:
		arr := []any{}
		for _, elem := range v {
			arr = append(arr, to.seriesJson(elem))
		}
		return arr
	case time.Time:
		if to.valueFormat.Timeformat.IsEpoch() {
			return to.valueFormat.Timeformat.Epoch(v)
		} else {
			return to.valueFormat.Timeformat.Format(v)
		}
	case float64:
		if to.valueFormat.Decimal == 0 {
			return int(v)
		} else if to.valueFormat.Decimal > 0 {
			return &util.JsonFloat{Value: v, Decimal: to.valueFormat.Decimal}
		} else {
			return v
		}
	default:
		return val
	}
}
