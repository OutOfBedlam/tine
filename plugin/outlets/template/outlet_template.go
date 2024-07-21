package template

import (
	"io"
	"os"
	gotmpl "text/template"

	"github.com/OutOfBedlam/tine/engine"
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
	ctx    *engine.Context
	tmpl   *gotmpl.Template
	writer io.WriteCloser
}

func (to *templateOutlet) Open() error {
	conf := to.ctx.Config()
	templates := conf.GetStringArray("templates", []string{})
	templateFiles := conf.GetStringArray("template_files", []string{})

	to.tmpl = gotmpl.New("template")
	var err error
	for _, t := range templates {
		to.tmpl, err = to.tmpl.Parse(t)
		if err != nil {
			return err
		}
	}
	if len(templateFiles) > 0 {
		to.tmpl, err = gotmpl.ParseFiles(templateFiles...)
		if err != nil {
			return err
		}
	}

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
	if to.writer != nil {
		to.writer.Close()
	}
	return nil
}

func (to *templateOutlet) Handle(recs []engine.Record) error {
	arr := []map[string]any{}
	for _, rec := range recs {
		fields := rec.Fields()
		obj := map[string]any{}
		for _, f := range fields {
			obj[f.Name] = f.Value
		}
		arr = append(arr, obj)
	}
	err := to.tmpl.Execute(to.writer, arr)
	if err != nil {
		return err
	}
	return nil
}
