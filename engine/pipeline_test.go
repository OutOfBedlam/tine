package engine_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"slices"
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
	_ "github.com/OutOfBedlam/tine/plugin/inlets/psutil"
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

func ExampleNewReader_withTypes() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1,1.2345", 
			"b,2,2.3456", 
			"c,3,3.4567",
		]
		format = "csv"
		fields = ["area", "ival", "fval"]
		types  = ["string", "int", "float"]
	[[flows.select]]
		includes = ["#*", "area", "ival", "fval"]
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
	// {"_in":"file","_ts":1721954797,"area":"a","fval":1.2345,"ival":1}
	// {"_in":"file","_ts":1721954797,"area":"b","fval":2.3456,"ival":2}
	// {"_in":"file","_ts":1721954797,"area":"c","fval":3.4567,"ival":3}
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

func TestMultiInlets(t *testing.T) {
	dsl := `
	[[inlets.cpu]]
		percpu = false
		interval = "1s"
		count = 2
	[[inlets.load]]
		loads = [1, 5]
		interval = "1s"
		count = 1
	[[flows.merge]]
		wait_limit = "1s"
	[[outlets.file]]
		path = "-"
		format = "json"
		decimal = 2
	`

	out := &bytes.Buffer{}
	pipeline, err := engine.New(engine.WithConfig(dsl), engine.WithWriter(out))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	dec := json.NewDecoder(out)
	obj := map[string]interface{}{}

	expectKeys := [][]string{
		{"_ts", "cpu.total_percent", "load.load1", "load.load5"},
		{"_ts", "cpu.total_percent"},
	}
	for i := 0; dec.Decode(&obj) == nil; i++ {
		t.Log(obj)
		keys := []string{}
		for k := range obj {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		require.Equal(t, expectKeys[i], keys)
		clear(obj)
	}
}

func ExampleContext_Inject() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[[inlets.args]]
	[[flows.select]]
		includes = ["**"]
	[[flows.inject]]
		id = "here"
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

	pipeline.Context().Inject("here", func(r []engine.Record) ([]engine.Record, error) {
		for i, rec := range r {
			r[i] = rec.AppendOrReplace(engine.NewField("msg", "hello world - here updated"))
		}
		return r, nil
	})

	// Simulate the command line arguments
	os.Args = []string{"command", "command-arg", "--", "msg=hello world"}
	// Run the pipeline
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"_in":"args","_ts":1721954797,"msg":"hello world - here updated"}
}
