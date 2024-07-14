package file

import (
	"io"
	"log/slog"
	"os"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "file",
		Factory: FileInlet,
	})
}

func FileInlet(ctx *engine.Context) engine.Inlet {
	return &fileInlet{ctx: ctx}
}

type fileInlet struct {
	ctx    *engine.Context
	reader *engine.Reader
	closer io.Closer
}

var _ = engine.PushInlet((*fileInlet)(nil))

func (fi *fileInlet) Open() error {
	path := fi.ctx.Config().GetString("path", "-")
	readerConf := fi.ctx.Config().GetConfig("reader", engine.Config{"format": "csv"})

	slog.Debug("inlet.file", "path", path, "reader", readerConf)

	var in io.Reader
	if f, err := os.Open(path); err != nil {
		return err
	} else {
		in = f
		fi.closer = f
	}

	r, err := engine.NewReader(
		engine.WithReader(in),
		engine.WithReaderConfig(readerConf),
	)
	if err != nil {
		return err
	}
	fi.reader = r
	return nil
}

func (fi *fileInlet) Close() error {
	if fi.reader != nil {
		return fi.reader.Close()
	}
	if fi.closer != nil {
		if err := fi.closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (si *fileInlet) Push(cb func([]engine.Record, error)) {
	for {
		r, err := si.reader.Read()
		cb(r, err)
		if err != nil {
			break
		}
	}
}
