package base

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"io"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "deflate",
		ContentEncoding: "deflate",
		Factory: func(w io.Writer) io.WriteCloser {
			ret, _ := flate.NewWriter(w, flate.BestSpeed)
			return ret
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "inflate",
		Factory: func(r io.Reader) io.ReadCloser {
			return flate.NewReader(r)
		},
	})
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "flate",
		ContentEncoding: "flate",
		Factory: func(w io.Writer) io.WriteCloser {
			ret, _ := flate.NewWriter(w, flate.BestSpeed)
			return ret
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "flate",
		Factory: func(r io.Reader) io.ReadCloser {
			return flate.NewReader(r)
		},
	})
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "gzip",
		ContentEncoding: "gzip",
		Factory: func(w io.Writer) io.WriteCloser {
			return gzip.NewWriter(w)
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "gzip",
		Factory: func(r io.Reader) io.ReadCloser {
			ret, _ := gzip.NewReader(r)
			return ret
		},
	})
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "lzw",
		ContentEncoding: "lzw",
		Factory: func(w io.Writer) io.WriteCloser {
			return lzw.NewWriter(w, lzw.LSB, 8)
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "lzw",
		Factory: func(r io.Reader) io.ReadCloser {
			return lzw.NewReader(r, lzw.LSB, 8)
		},
	})
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "zlib",
		ContentEncoding: "zlib",
		Factory: func(w io.Writer) io.WriteCloser {
			return zlib.NewWriter(w)
		},
	})
	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "zlib",
		Factory: func(r io.Reader) io.ReadCloser {
			ret, _ := zlib.NewReader(r)
			return ret
		},
	})
}
