package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	"github.com/OutOfBedlam/tine/plugin/inlets/psutil"
	"github.com/OutOfBedlam/tine/plugin/outlets/file"
)

func main() {
	// Create a pipeline
	pipeline, err := engine.New(engine.WithName("custom_flow"))
	if err != nil {
		panic(err)
	}

	interval := 3 * time.Second
	cpuNum := runtime.NumCPU()

	// Add inlet for cpu usage
	conf := engine.NewConfig().Set("loads", []int{1}).Set("interval", interval)
	pipeline.AddInlet("load", psutil.LoadInlet(pipeline.Context().WithConfig(conf)))

	// Add outlet printing to stdout '-'
	conf = engine.NewConfig().Set("path", "-").Set("timeformat", "2006-01-02 15:04:05")
	pipeline.AddOutlet("file", file.FileOutlet(pipeline.Context().WithConfig(conf)))

	// Add your custom flow function.
	custom := func(recs []engine.Record) ([]engine.Record, error) {
		ret := make([]engine.Record, 0, len(recs))
		for _, r := range recs {
			if field := r.Field("load1"); field != nil {
				load, _ := field.Value.Float64()
				percent := (load * 100) / float64(cpuNum)
				r = r.Append(engine.NewField("load1_percent", fmt.Sprintf("%.1f%%", percent)))
			}
			ret = append(ret, r)
		}
		return ret, nil
	}
	pipeline.AddFlow("custom", engine.FlowWithFunc(custom, engine.WithFlowFuncParallelism(1)))

	// Start the pipeline
	pipeline.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the pipeline
	pipeline.Stop()
}
