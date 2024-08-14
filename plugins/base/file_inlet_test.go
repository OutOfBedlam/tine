package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleFileInlet() {
	dsl := `
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// a,1
	// b,2
	// c,3
}

func ExampleFileInlet_file() {
	dsl := `
	[[inlets.file]]
		path = "testdata/testdata.csv"
		format = "csv"
		fields = ["line", "name", "time", "value"]
		types  = ["int", "string", "time", "float"]
	[[outlets.file]]
		path = "-"
		format = "json"
		decimal = 2
	`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"line":1,"name":"key1","time":1722642405,"value":1.23}
	// {"line":2,"name":"key2","time":1722642406,"value":2.35}
}

func ExampleFileInlet_fields() {
	dsl := `
	[[inlets.file]]
		data = [
			"1,key1,1722642405,1.234",
			"2,key2,1722642406,2.345",
		]
		format = "csv"
		fields = ["line", "name", "time", "value"]
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
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// {"line":"1","name":"key1","time":"1722642405","value":"1.234"}
	// {"line":"2","name":"key2","time":"1722642406","value":"2.345"}
}
