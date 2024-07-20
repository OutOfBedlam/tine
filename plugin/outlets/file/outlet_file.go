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
	ctx    *engine.Context
	writer *engine.Writer
	closer io.Closer
}

func (fo *fileOutlet) Open() error {
	path := fo.ctx.Config().GetString("path", "-")
	writerConf := fo.ctx.Config().GetConfig("writer", engine.Config{"format": "csv"})
	fo.ctx.LogDebug("outlet.file", "path", path, "format", writerConf.GetString("format", "?"))

	var out io.Writer
	if w := fo.ctx.Writer(); w != nil {
		out = w
	} else {
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
	w, err := engine.NewWriter(out,
		engine.WithWriterConfig(writerConf),
	)
	if err != nil {
		return err
	}
	fo.ctx.SetContentType(w.ContentType)
	fo.ctx.SetContentEncoding(w.ContentEncoding)
	fo.writer = w
	return nil
}

func (fo *fileOutlet) Close() error {
	if fo.writer != nil {
		if err := fo.writer.Close(); err != nil {
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
	return fo.writer.Write(recs)
}
