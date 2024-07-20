package main

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

var config = `
[log]
	level = "error"

[[inlets.exec]]
	commands = ["echo", "Hello, World!"]
	count = 1

[[outlets.file]]
	path = "-"
	[[outlets.file.writer]]
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
