package engine_test

import (
	"compress/zlib"
	"io"
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/stretchr/testify/require"
)

func TestCompressor(t *testing.T) {
	engine.RegisterCompressor(&engine.Compressor{
		Name:            "test-zlib",
		ContentEncoding: "test-zlib",
		Factory: func(w io.Writer) io.WriteCloser {
			return zlib.NewWriter(w)
		},
	})

	engine.RegisterDecompressor(&engine.Decompressor{
		Name: "test-zlib",
		Factory: func(r io.Reader) io.ReadCloser {
			ret, _ := zlib.NewReader(r)
			return ret
		},
	})
	names := engine.CompressorNames()
	require.Equal(t, []string{"deflate", "flate", "gzip", "lzw", "test-zlib", "zlib"}, names)

	names = engine.DecompressorNames()
	require.Equal(t, []string{"flate", "gzip", "inflate", "lzw", "test-zlib", "zlib"}, names)

	enc := engine.GetCompressor("test-zlib")
	require.NotNil(t, enc)
	require.Equal(t, "test-zlib", enc.Name)

	dec := engine.GetDecompressor("test-zlib")
	require.NotNil(t, dec)
	require.Equal(t, "test-zlib", dec.Name)

	engine.UnregisterCompressor("test-zlib")
	engine.UnregisterDecompressor("test-zlib")

	names = engine.CompressorNames()
	require.Equal(t, []string{"deflate", "flate", "gzip", "lzw", "zlib"}, names)

	names = engine.DecompressorNames()
	require.Equal(t, []string{"flate", "gzip", "inflate", "lzw", "zlib"}, names)
}
