package base_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleDumpFlow() {
	recipe := `
		name="pipeline-1"
		[log]
			path = "-"
			level = "warn"
			no_color = true
			timeformat = "no-time-for-test"
		[[inlets.file]]
			data = [
				"a,100",
				"b,200",
			]
		[[inlets.file.flows.dump]]
			level = "warn"
		[[flows.dump]]
			level = "error"
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
	// no-time-for-test WRN pipeline pipeline-1 flow-dump rec=1/2 0=a 1=100
	// no-time-for-test WRN pipeline pipeline-1 flow-dump rec=2/2 0=b 1=200
	// no-time-for-test ERR pipeline pipeline-1 flow-dump rec=1/2 0=a 1=100
	// no-time-for-test ERR pipeline pipeline-1 flow-dump rec=2/2 0=b 1=200
	// {"0":"a","1":"100"}
	// {"0":"b","1":"200"}
}
