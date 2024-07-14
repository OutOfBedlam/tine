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
	// Create engine
	eng, err := engine.New(engine.WithName("example"))
	if err != nil {
		panic(err)
	}

	interval := 3 * time.Second

	// Add inlet for cpu usage
	conf := engine.NewConfig().Set("percpu", false).Set("interval", interval)
	eng.AddInlet("cpu", psutil.CpuInlet(eng.Context().WithConfig(conf)))

	// Add outlet printing to stdout '-'
	conf = engine.NewConfig().Set("path", "-").Set("interval", interval)
	eng.AddOutlet(
		file.FileOutlet(eng.Context().WithConfig(conf)))

	// Add your custom input function.
	custom := func() ([]engine.Record, error) {
		result := []engine.Record{
			engine.NewRecord(
				engine.NewStringField("name", "random"),
				engine.NewFloatField("value", rand.Float64()*100),
			),
		}
		return result, nil
	}
	eng.AddInlet("custom", engine.InletWithPullFunc(custom, interval))

	// Start the engine
	go eng.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the engine
	eng.Stop()
}
