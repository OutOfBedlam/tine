package base_test

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
)

func ExampleDamperFlow() {
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
				"c,300",
				"d,400",
			]
		[[flows.damper]]
			buffer_limit = 2
		[[flows.select]]
			includes = ["#_ts", "*"]
		[[outlets.file]]
			path = "-"
			format = "json"
	`
	seq := int64(0)
	engine.Now = func() time.Time { seq++; return time.Unix(1721954797+seq, 0) }
	pipe, err := engine.New(engine.WithConfig(recipe))
	if err != nil {
		panic(err)
	}
	err = pipe.Run()
	if err != nil {
		panic(err)
	}

	// Output:
	// {"0":"a","1":"100","_ts":1721954798}
	// {"0":"b","1":"200","_ts":1721954799}
	// {"0":"c","1":"300","_ts":1721954800}
	// {"0":"d","1":"400","_ts":1721954801}
}
