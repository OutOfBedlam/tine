package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/containerd/console"
	"github.com/spf13/cobra"

	_ "github.com/OutOfBedlam/tine/plugin/all"
	_ "github.com/OutOfBedlam/tine/x"
)

func NewCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	if runtime.GOOS == "windows" {
		console.ConsoleFromFile(os.Stdin) //nolint:errcheck
	}

	rootCmd := &cobra.Command{
		Use:           "tine [command] [flags] [args]",
		Short:         "TINE is not ETL, but pipeline runner",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print(cmd.UsageString())
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all plugins",
		RunE:  ListHandler,
	}

	runCmd := &cobra.Command{
		Use:   "run [flags] FILE [, FILE ...]",
		Short: "Run tine pipelines from the specified one or more files",
		Args:  cobra.MinimumNArgs(1),
		RunE:  RunHandler,
	}
	runCmd.Flags().String("pid", "", "write PID to the `<path>` file")

	rootCmd.AddCommand(
		runCmd,
		listCmd,
	)
	return rootCmd

}

func ListHandler(cmd *cobra.Command, args []string) error {
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
	return nil
}

func RunHandler(cmd *cobra.Command, args []string) error {
	if pos := cmd.ArgsLenAtDash(); pos >= 0 {
		// passthroughArgs := args[pos+1:]
		args = args[0:pos]
	}
	if err := cmd.ParseFlags(args); err != nil {
		return err
	}

	optPid, _ := cmd.Flags().GetString("pid")
	optPid, err := filepath.Abs(optPid)
	if err != nil {
		return err
	}
	pipelineConfigs := []string{}
	for _, arg := range args {
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
	if optPid != "" {
		pfile, _ := os.OpenFile(optPid, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		pfile.WriteString(fmt.Sprintf("%d", os.Getpid()))
		pfile.Close()
		defer func() {
			os.Remove(optPid)
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
	return nil
}
