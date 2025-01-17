package main

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/all"
)

var config = `
[[inlets.exec]]
	commands = ["echo", "Hello, World!"]
	count = 1

[[flows.select]]
	includes = ["#_ts", "#_in", "*"]

[[outlets.file]]
	path = "-"
	format = "json"
`

func main() {
	// Create engine
	engine, err := engine.New(
		engine.WithName("helloworld"),
		engine.WithConfig(config),
	)
	if err != nil {
		panic(err)
	}

	// Execute engine
	err = engine.Run()
	if err != nil {
		panic(err)
	}

	// Stop engine
	engine.Stop()
}
