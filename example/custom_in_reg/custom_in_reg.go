package main

import (
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/psutil"
)

const config = `
[[inlets.cpu]]
	interval = "3s"
	percpu = false

[[inlets.random]]
	interval = "3s"

[[outlets.file]]
	path = "-"
	format = "csv"  # output format "csv" | "json"
	decimal = 2
`

func main() {
	// Register custom inlet as named "random".
	// This inlet will generate records that has a "random" field.
	engine.RegisterInlet(&engine.InletReg{
		Name: "random",
		Factory: func(ctx *engine.Context) engine.Inlet {
			interval := ctx.Config().GetDuration("interval", 10*time.Second)
			return engine.InletWithFunc(customInletFunc, engine.WithInterval(interval))
		},
	})

	// Create a pipeline
	pipeline, err := engine.New(engine.WithConfig(config))
	if err != nil {
		panic(err)
	}

	// Start the pipeline
	pipeline.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the pipeline
	pipeline.Stop()
}

func customInletFunc() ([]engine.Record, error) {
	result := []engine.Record{
		engine.NewRecord(
			engine.NewField("text", "hello world"),
			engine.NewField("random", rand.Float64()*100),
		),
	}
	return result, nil
}
