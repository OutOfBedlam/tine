package csv_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleCSVEncoder() {
	dsl := `
	[log]
		level = "warn"
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
		fields = ["area"]
	[[outlets.file]]
		path = "-"
		format = "csv"
		fields = ["_in", "0", "1"]
	`
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// file,a,1
	// file,b,2
	// file,c,3
}
