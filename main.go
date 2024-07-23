package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/OutOfBedlam/tine/engine"

	_ "github.com/OutOfBedlam/tine/plugin/all"
	_ "github.com/OutOfBedlam/tine/x"
)

var usageStr = `
Usage: tine [options] <pipeline.toml> ...

Run tine with the specified pipeline configuration file.
Multiple pipeline configuration files can be specified.

Options:
	--help, -h              Show this help message
	--pid                   PID file path
	--list                  List all plugins
`

func usage() {
	fmt.Printf("%s\n", strings.ReplaceAll(usageStr, "\t", "    "))
	os.Exit(0)
}

func listPlugins() int {
	fmt.Println("Input: data -> [inlet] -> [decompress] -> [decoder] -> records")
	fmt.Println("  Decoders   ", strings.Join(engine.DecoderNames(), ","))
	fmt.Println("  Decompress ", strings.Join(engine.DecompressorNames(), ","))
	fmt.Println("  Inlets     ", strings.Join(engine.InletNames(), ","))
	fmt.Println("")
	fmt.Println("Output: records -> [encoder] -> [compress] -> [outlet] -> data")
	fmt.Println("  Encoders   ", strings.Join(engine.EncoderNames(), ","))
	fmt.Println("  Compress   ", strings.Join(engine.CompressorNames(), ","))
	fmt.Println("  Outlets    ", strings.Join(engine.OutletNames(), ","))
	fmt.Println("")
	fmt.Println("Flows        ", strings.Join(engine.FlowNames(), ","))
	return 0
}

func main() {
	optPid := flag.String("pid", "", "pid file path")
	optList := flag.Bool("list", false, "list all plugins")

	flag.Usage = usage
	flag.Parse()

	if *optList {
		os.Exit(listPlugins())
	}

	pipelineConfigs := []string{}
	for _, arg := range flag.Args() {
		if arg == "--" {
			break
		}
		if _, err := os.Stat(arg); err != nil {
			fmt.Println("pipeline file not found:", arg)
			os.Exit(1)
		}
		pipelineConfigs = append(pipelineConfigs, arg)
	}

	if len(pipelineConfigs) == 0 {
		fmt.Println("no pipeline file specified")
		os.Exit(1)
	}

	pipelines := make([]*engine.Pipeline, 0, len(pipelineConfigs))
	for _, pc := range pipelineConfigs {
		p, err := engine.New(engine.WithConfigFile(pc))
		if err != nil {
			fmt.Println("failed to parse pipeline file:", err)
			os.Exit(1)
		}
		pipelines = append(pipelines, p)
	}

	// PID file
	if *optPid != "" {
		pfile, _ := os.OpenFile(*optPid, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		pfile.WriteString(fmt.Sprintf("%d", os.Getpid()))
		pfile.Close()
		defer func() {
			os.Remove(*optPid)
		}()
	}

	// run pipelines
	waitGroup := sync.WaitGroup{}
	for _, pipe := range pipelines {
		waitGroup.Add(1)
		go func() {
			pipe.Run()
			waitGroup.Done()
		}()
	}

	waitCh := make(chan struct{})
	go func() {
		waitGroup.Wait()
		waitCh <- struct{}{}
	}()

	// wait Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-interrupt:
		for _, pipe := range pipelines {
			pipe.Stop()
		}
		waitGroup.Wait()
	case <-waitCh:
	}
	os.Exit(0)
}
