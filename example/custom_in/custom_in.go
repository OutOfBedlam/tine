package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	"github.com/OutOfBedlam/tine/plugin/inlets/psutil"
	"github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func main() {
	// Create pipeline
	pipeline, err := engine.New(engine.WithName("custom_in"))
	if err != nil {
		panic(err)
	}

	interval := 3 * time.Second

	// Add inlet for cpu usage
	conf := engine.NewConfig().Set("percpu", false).Set("interval", interval)
	pipeline.AddInlet("cpu", psutil.CpuInlet(pipeline.Context().WithConfig(conf)))

	// Add outlet printing to stdout '-'
	conf = engine.NewConfig().Set("path", "-").Set("decimal", 2)
	pipeline.AddOutlet("file", file.FileOutlet(pipeline.Context().WithConfig(conf)))

	// Add your custom input function.
	custom := func() ([]engine.Record, error) {
		result := []engine.Record{
			engine.NewRecord(
				engine.NewField("name", "random"),
				engine.NewField("value", rand.Float64()*100),
			),
		}
		return result, nil
	}
	pipeline.AddInlet("custom", engine.InletWithFunc(custom, engine.WithInterval(interval)))

	// Start the pipeline
	pipeline.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the pipeline
	pipeline.Stop()
}
