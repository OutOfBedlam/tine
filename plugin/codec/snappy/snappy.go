package snappy

import (
	"io"
	"os"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/golang/snappy"
)

func init() {
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "snappy",
		ContentEncoding: "snappy",
		Factory: func(w io.Writer) io.WriteCloser {
			return snappy.NewBufferedWriter(os.Stdout)
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "snappy",
		Factory: func(r io.Reader) io.ReadCloser {
			ret := snappy.NewReader(r)
			return io.NopCloser(ret)
		},
	})
}
