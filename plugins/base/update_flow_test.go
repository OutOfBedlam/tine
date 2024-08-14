package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleUpdateFlow() {
	dsl := `
	[[inlets.file]]
		data = [
			"James,1,1.23,true", 
			"Jane,2,2.34,false", 
			"Scott,3,3.45,true",
		]
		fields = ["my_name", "my_int", "my_float", "flag"]
		format = "csv"
	[[flows.update]]
		set = [
			{ field = "my_name", name = "new_name" },
			{ field = "my_int", value = 10 },
			{ field = "my_float", value = 9.87, name = "new_float" },
			{ field = "flag", value = true, name = "new_flag" },
			{ tag = "_in", value = "mine" },
		]
	[[flows.select]]
		includes = ["#_in", "*"]
	[[outlets.file]]
		format = "json"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
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
	// {"_in":"mine","my_int":"10","new_flag":true,"new_float":9.87,"new_name":"James"}
	// {"_in":"mine","my_int":"10","new_flag":true,"new_float":9.87,"new_name":"Jane"}
	// {"_in":"mine","my_int":"10","new_flag":true,"new_float":9.87,"new_name":"Scott"}
}
