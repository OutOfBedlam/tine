package exec_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/exec"
)

func ExampleExecFlow_sample() {
	// This example demonstrates how to use the exec inlet to run a command and
	dsl := `
	[[inlets.file]]
		data = [
			"a,1",
			"b,2",
		]
		format = "csv"
	[[flows.exec]]
		commands = ["sh", "-c", "echo hello $FOO $FIELD_0 $FIELD_1"]
		environments = ["FOO=BAR"]
		trim_space = true
		ignore_error = true
	[[flows.select]]
		includes= ["#_ts", "stdout"]
	[[outlets.file]]
		path = "-"
		format = "json"
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
	// {"_ts":1721954798,"stdout":"hello BAR a 1"}
	// {"_ts":1721954799,"stdout":"hello BAR b 2"}
}
