package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/psutil"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
)

const config = `
[[inlets.load]]
	loads = [ 1 ]
	interval = "3s"

[[flows.custom]]

[[outlets.file]]
	path = "-"
	format = "csv"  # set output format "csv" and also "json" is available
	timeformat = "2006-01-02 15:04:05"
`

func main() {
	// Register custom flow as named "custom".
	// This flow will add a new field "load1_percent" to each record.
	engine.RegisterFlow(&engine.FlowReg{
		Name: "custom",
		Factory: func(ctx *engine.Context) engine.Flow {
			return engine.FlowWithFunc(customFlowFunc)
		},
	})

	// Create a pipeline
	pipeline, err := engine.New(engine.WithConfig(config))
	if err != nil {
		panic(err)
	}

	// Start the pipeline
	go pipeline.Start()

	// wait Ctrl+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop the pipeline
	pipeline.Stop()
}

func customFlowFunc(recs []engine.Record) ([]engine.Record, error) {
	cpuNum := runtime.NumCPU()
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
