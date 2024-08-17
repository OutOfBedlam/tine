package monad_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/monad"
)

func ExampleFilterFlow() {
	recipe := `
	[log]
		path = "-"
		level = "warn"
		no_color = true
		timeformat = "no-time-for-test"
	[[inlets.file]]
		data = [
			"a,100",
			"b,200",
			"c,300",
		]
		format = "csv"
		fields = ["name", "value"]
		types  = ["string", "int"]
	[[flows.filter]]
		predicate = "${ value } > 100"
	[[outlets.file]]
		path = "-"
		format = "json"
`
	pipe, err := engine.New(engine.WithConfig(recipe))
	if err != nil {
		panic(err)
	}
	err = pipe.Run()
	if err != nil {
		panic(err)
	}

	// Output:
	// {"name":"b","value":200}
	// {"name":"c","value":300}
}
