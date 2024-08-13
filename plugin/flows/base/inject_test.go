package base_test

import (
	"os"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/args"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleInjectFlow() {
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
