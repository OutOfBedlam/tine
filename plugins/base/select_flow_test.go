package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleSelectFlow() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
	[[flows.select]]
		includes = ["**", "not-exist", "#_ts", "1"]
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }
	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Run the pipeline
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// file,1721954798,a,1,,1721954798,1
	// file,1721954799,b,2,,1721954799,2
	// file,1721954800,c,3,,1721954800,3
}

func ExampleSelectFlow_tag() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
		fields = ["area", "ival"]
		types = ["string", "int"]
	[[flows.select]]
		includes = ["#_in", "#non_exist", "*"]
	[[outlets.file]]
		format = "json"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }
	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Run the pipeline
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"_in":"file","area":"a","ival":1,"non_exist":null}
	// {"_in":"file","area":"b","ival":2,"non_exist":null}
	// {"_in":"file","area":"c","ival":3,"non_exist":null}
}
