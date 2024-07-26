package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/exec"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleMergeFlow() {
	// This example demonstrates how to use the merge flow.
	dsl := `
	[log]
		level = "warn"
	[[inlets.file]]
		data = [
			"a,1",
		]
		format = "csv"
	[[inlets.exec]]
		commands = ["echo", "hello world"]
		trim_space = true
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
	// [{"_ts":1721954797,"exec.stdout":"hello world","file.0":"a","file.1":"1"}]
}
