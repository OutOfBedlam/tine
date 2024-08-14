package base_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	"github.com/stretchr/testify/require"
)

func TestCompressGzip(t *testing.T) {
	// This example demonstrates how to use the compress flow.
	dsl := `
	[[inlets.file]]
		data = [
			"a,1",
		]
		format = "csv"
	[[outlets.file]]
		path = "-"
		format = "json"
		compress = "gzip"
	`
	// Make the output time deterministic. so we can compare it.
	// This line is not needed in production code.
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Create a new engine.
	out := &bytes.Buffer{}
	pipeline, err := engine.New(engine.WithConfig(dsl), engine.WithWriter(out))
	if err != nil {
		panic(err)
	}
	// Run the engine.
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	r, _ := gzip.NewReader(out)
	result, _ := io.ReadAll(r)
	require.Equal(t, `{"0":"a","1":"1"}`, strings.TrimSpace(string(result)))
}
