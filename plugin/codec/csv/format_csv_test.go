package csv_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleCSVEncoder() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
	[[flows.select]]
		includes = ["#*", "1", "0"]
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	// Mock the current time
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }

	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// file,1721954798,1,a
	// file,1721954799,2,b
	// file,1721954800,3,c
}
