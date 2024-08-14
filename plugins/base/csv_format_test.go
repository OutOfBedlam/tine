package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleCSVEncoder() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1,1.234,true,2024/08/09 16:01:02", 
			"b,2,2.345,false,2024/08/09 16:03:04", 
			"c,3,3.456,true,2024/08/09 16:05:06",
		]
		format = "csv"
		timeformat = "2006/01/02 15:04:05"
		tz = "UTC"
		fields = ["area","ival","fval","bval","tval"]
		types  = ["string", "int", "float", "bool", "time"]
	[[flows.select]]
		includes = ["#*", "ival", "area", "ival", "fval", "bval", "tval"]
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
	// file,1721954798,1,a,1,1.234,true,1723219262
	// file,1721954799,2,b,2,2.345,false,1723219384
	// file,1721954800,3,c,3,3.456,true,1723219506
}
