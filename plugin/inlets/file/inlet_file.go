package file

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/OutOfBedlam/tine/util"
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

var _ = engine.Inlet((*fileInlet)(nil))

func (fi *fileInlet) Open() error {
	path := fi.ctx.Config().GetString("path", "")
	readerConf := fi.ctx.Config().GetConfig("reader", engine.Config{"format": "csv"})
	data := fi.ctx.Config().GetStringSlice("data", nil)
	slog.Debug("inlet.file", "path", path, "data", util.FormatCount(len(data), util.CountUnitLines), "reader", readerConf)
	if path == "" && len(data) == 0 {
		return fmt.Errorf("no path or data specified")
	}

	var in io.Reader
	if len(data) > 0 {
		in = bytes.NewBufferString(strings.Join(data, "\n"))
	} else {
		if f, err := os.Open(path); err != nil {
			return err
		} else {
			in = f
			fi.closer = f
		}
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

func (si *fileInlet) Process(cb engine.InletNextFunc) {
	for {
		r, err := si.reader.Read()
		cb(r, err)
		if err != nil {
			break
		}
	}
}
