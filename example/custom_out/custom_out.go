package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	"github.com/OutOfBedlam/tine/plugin/inlets/psutil"
)

func main() {
	// Create a pipeline
	pipeline, _ := engine.New(engine.WithName("custom_out"))

	// Add inlet getting cpu usage
	ctx := pipeline.Context().WithConfig(map[string]any{
		"percpu":   false,
		"interval": 3 * time.Second,
	})
	pipeline.AddInlet("cpu", psutil.CpuInlet(ctx))

	// Add outlet printing to stdout in custom format
	pipeline.AddOutlet("custom", engine.OutletWithFunc(func(recs []engine.Record) error {
		for _, r := range recs {
			if field := r.Field("total_percent"); field != nil {
				cpu, _ := field.Value.Float64()
				fmt.Printf("cpu usage: %.2f%%\n", cpu)
			}
		}
		return nil
	}))

	// Start the pipeline
	go pipeline.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the pipeline
	pipeline.Stop()
}
