package exec_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/exec"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleExecInlet() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[[inlets.exec]]
		commands = ["echo", "hello", "world"]
		trim_space = true
		count = 1
		ignore_error = true
	[[flows.select]]
		includes= ["#_ts", "stdout"]
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }
	// Build pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// 1721954798,hello world
}
