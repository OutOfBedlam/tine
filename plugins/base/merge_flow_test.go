package base_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/exec"
	_ "github.com/OutOfBedlam/tine/plugins/psutil"
)

func ExampleMergeFlow() {
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
	// {"_ts":1721954797,"exec_stdout":"hello world","file_0":"a","file_1":"1"}
}

func TestExampleMergeFlow(t *testing.T) {
	// This example demonstrates how to use the merge flow.
	dsl := `
	[[inlets.load]]
		interval = "1s"
		count = 3
		loads = [1]
	[[inlets.cpu]]
		interval = "1s"
		count = 3
		percpu = false
		total = true
	[[flows.merge]]
		wait_limit = "1.5s"
	[[outlets.file]]
		path = "-"
		format = "json"
	`
	// Make the output time deterministic. so we can compare it.
	// This line is not needed in production code.
	seq := int64(0)
	engine.Now = func() time.Time { seq++; return time.Unix(1721954797+(seq/4), 0) }
	// Create a new engine.
	buff := &bytes.Buffer{}
	pipeline, err := engine.New(engine.WithConfig(dsl), engine.WithWriter(buff))
	if err != nil {
		panic(err)
	}
	// Run the engine.
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	pipeline.Stop()

	dec := json.NewDecoder(buff)
	for i := 0; i < 3; i++ {
		rec := map[string]interface{}{}
		if err := dec.Decode(&rec); err != nil {
			t.Fatal(err)
		}
		if ts := rec["_ts"]; ts == nil {
			t.Fatal("missing _ts")
		} else {
			if int64(ts.(float64)) != 1721954797+int64(i) {
				t.Fatalf("unexpected _ts: %v", ts)
			}
		}
		if rec["load_load1"] == nil {
			t.Fatal("missing load_0")
		}
		if rec["cpu_total_percent"] == nil {
			t.Fatal("missing cpu_total")
		}
	}
}
