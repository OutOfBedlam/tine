package args_test

import (
	"os"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/args"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleArgsInlet() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[[inlets.args]]
	[[outlets.file]]
		path = "-"
		format = "csv"
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
	// hello world
}
