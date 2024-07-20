package file

import (
	"io"
	"os"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "file",
		Factory: FileOutlet,
	})
}

func FileOutlet(ctx *engine.Context) engine.Outlet {
	return &fileOutlet{ctx: ctx}
}

type fileOutlet struct {
	*engine.Writer
	ctx    *engine.Context
	closer io.Closer
}

func (fo *fileOutlet) Open() error {
	var out io.Writer
	if w := fo.ctx.Writer(); w != nil {
		out = w
	} else {
		path := fo.ctx.Config().GetString("path", "-")
		if path == "" {
			out = io.Discard
		} else if path == "-" {
			out = io.Writer(os.Stdout)
		} else {
			if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err != nil {
				fo.ctx.LogError("failed to open file", "path", path, "error", err.Error())
			} else {
				out = f
				fo.closer = f
			}
		}
	}

	w, err := engine.NewWriter(out, fo.ctx.Config())
	if err != nil {
		return err
	}
	fo.ctx.SetContentType(w.ContentType)
	fo.ctx.SetContentEncoding(w.ContentEncoding)
	fo.Writer = w
	return nil
}

func (fo *fileOutlet) Close() error {
	if fo.Writer != nil {
		if err := fo.Writer.Close(); err != nil {
			return err
		}
	}
	if fo.closer != nil {
		if err := fo.closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (fo *fileOutlet) Handle(recs []engine.Record) error {
	return fo.Writer.Write(recs)
}
