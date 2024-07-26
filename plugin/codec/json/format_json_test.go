package json_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func ExampleJSONEncoder() {
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
	[[flows.set_field]]
		set = { _ts = 1721954797 }
	[[outlets.file]]
		path = "-"
		format = "json"
	`
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// [{"0":"a","1":"1","_in":"file","_ts":"1721954797"},{"0":"b","1":"2","_in":"file","_ts":"1721954797"},{"0":"c","1":"3","_in":"file","_ts":"1721954797"}]
}
