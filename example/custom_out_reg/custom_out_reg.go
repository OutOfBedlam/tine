package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/psutil"
)

const config = `
[[inlets.cpu]]
	interval = "3s"
	percpu = false

[[outlets.custom]]
`

func main() {
	// Register custom outlet as named "custom".
	// This outlet will print out "total_percent" field with "%" suffix.
	engine.RegisterOutlet(&engine.OutletReg{
		Name: "custom",
		Factory: func(ctx *engine.Context) engine.Outlet {
			return engine.OutletWithFunc(customOutletFunc)
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

func customOutletFunc(recs []engine.Record) error {
	for _, r := range recs {
		if field := r.Field("total_percent"); field != nil {
			cpu, _ := field.Value.Float64()
			fmt.Printf("cpu usage: %.2f%%\n", cpu)
		}
	}
	return nil
}
