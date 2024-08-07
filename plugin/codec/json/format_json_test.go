package json_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleJSONEncoder() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1,1.2345", 
			"b,2,2.3456", 
			"c,3,3.4567",
		]
		format = "csv"
		fields = ["area", "ival", "fval"]
		types  = ["string", "int", "float"]
	[[flows.select]]
		includes = ["#*", "area", "ival", "fval"]
	[[outlets.file]]
		path = "-"
		format = "json"
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
	// {"_in":"file","_ts":1721954798,"area":"a","fval":1.2345,"ival":1}
	// {"_in":"file","_ts":1721954799,"area":"b","fval":2.3456,"ival":2}
	// {"_in":"file","_ts":1721954800,"area":"c","fval":3.4567,"ival":3}
}

func ExampleJSONEncoder_decimal() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1,1.2345", 
			"b,2,2.3456", 
			"c,3,3.4567",
		]
		format = "csv"
		fields = ["area", "ival", "fval"]
		types  = ["string", "int", "float"]
	[[flows.select]]
		includes = ["#*", "area", "ival", "fval"]
	[[outlets.file]]
		path = "-"
		format = "json"
		decimal = 2
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
	// {"_in":"file","_ts":1721954798,"area":"a","fval":1.23,"ival":1}
	// {"_in":"file","_ts":1721954799,"area":"b","fval":2.35,"ival":2}
	// {"_in":"file","_ts":1721954800,"area":"c","fval":3.46,"ival":3}
}
