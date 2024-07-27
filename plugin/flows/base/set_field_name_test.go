package base_test

import (
	"os"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/args"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleSetFieldNameFlow() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[log]
		level = "warn"
	[[inlets.args]]
	[[flows.set_field_name]]
		prefix = "pre_"
		suffix = "_suf"
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
	// [{"_in":"args","_ts":1721954797,"pre_msg_suf":"hello world"}]
}
