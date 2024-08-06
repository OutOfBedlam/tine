package engine_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"compress/gzip"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/compress"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/args"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/exec"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
	"github.com/stretchr/testify/require"
)

func ExampleNew() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[[inlets.args]]
	[[flows.update]]
		set = [
			{ field = "msg", name_format = "pre_%s_suf" },
		]
	[[flows.select]]
		includes = ["**"]
	[[outlets.file]]
		path = "-"
		format = "json"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Build pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Simulate the command line arguments
	os.Args = []string{"command", "command-arg", "--", "msg=hello world"}
	// Run the pipeline
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"_in":"args","_ts":1721954797,"pre_msg_suf":"hello world"}
}

func ExampleNew_multi_inlets() {
	// This example demonstrates how to use the merge flow.
	dsl := `
	[[inlets.file]]
		data = [
			"a,1",
		]
		format = "csv"
	[[inlets.exec]]
		commands = ["echo", "hello world"]
		count = 1
		trim_space = true
		ignore_error = true
	[[flows.merge]]
		wait_limit = "1s"
	[[outlets.file]]
		path = "-"
		format = "json"
	`
	// Make the output time deterministic. so we can compare it.
	// This line is not needed in production code.
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Create a new engine.
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Run the engine.
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"_ts":1721954797,"exec.stdout":"hello world","file.0":"a","file.1":"1"}
}

func TestCompressJson(t *testing.T) {
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
