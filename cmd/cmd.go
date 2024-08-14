package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/OutOfBedlam/tine/tools/pipeviz"
	"github.com/containerd/console"
	"github.com/spf13/cobra"

	_ "github.com/OutOfBedlam/tine/plugins/all"
	_ "github.com/OutOfBedlam/tine/x"
)

func NewCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	if runtime.GOOS == "windows" {
		console.ConsoleFromFile(os.Stdin) //nolint:errcheck
	}

	rootCmd := &cobra.Command{
		Use:           "tine [command] [flags] [args]",
		Short:         "TINE is not ETL",
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of TINE",
		Run: func(cmd *cobra.Command, args []string) {
			short, _ := cmd.Flags().GetBool("short")
			if short {
				fmt.Println(engine.DisplayVersion())
			} else {
				fmt.Println(engine.VersionString())
			}
		},
	}
	versionCmd.Flags().BoolP("short", "s", false, "print the version only")

	graphCmd := &cobra.Command{
		Use:   "graph",
		Short: "Generate a graph of the pipeline",
		RunE:  GraphHandler,
	}
	graphCmd.Flags().StringP("output", "o", "-", "output file name")

	runCmd := &cobra.Command{
		Use:   "run [flags] FILE [, FILE ...]",
		Short: "Run tine pipelines from the specified one or more files",
		Args:  cobra.MinimumNArgs(1),
		RunE:  RunHandler,
	}
	runCmd.Flags().String("pid", "", "write PID to the `<path>` file")

	rootCmd.AddCommand(
		runCmd,
		graphCmd,
		listCmd,
		versionCmd,
	)
	return rootCmd
}

func ListHandler(cmd *cobra.Command, args []string) error {
	_split := func(names []string) string {
		slices.Sort(names)
		lines := []string{}
		sb := strings.Builder{}
		for i, name := range names {
			if i > 0 {
				sb.WriteString(", ")
			}
			if sb.Len()+len(name) > 60 {
				lines = append(lines, sb.String())
				sb.Reset()
			}
			sb.WriteString(name)
		}
		if sb.Len() > 0 {
			lines = append(lines, sb.String())
		}
		return strings.Join(lines, "\n              ")
	}
	fmt.Println("Input: data -> [inlet] -> [decompress] -> [decoder] -> records")
	fmt.Println("  Decoders   ", _split(engine.DecoderNames()))
	fmt.Println("  Decompress ", _split(engine.DecompressorNames()))
	fmt.Println("  Inlets     ", _split(engine.InletNames()))
	fmt.Println("")
	fmt.Println("Flows: records -> [flow ...] -> records")
	fmt.Println("  Flows      ", _split(engine.FlowNames()))
	fmt.Println("")
	fmt.Println("Output: records -> [encoder] -> [compress] -> [outlet] -> data")
	fmt.Println("  Encoders   ", _split(engine.EncoderNames()))
	fmt.Println("  Compress   ", _split(engine.CompressorNames()))
	fmt.Println("  Outlets    ", _split(engine.OutletNames()))
	fmt.Println("")
	return nil
}

func GraphHandler(cmd *cobra.Command, args []string) error {
	if pos := cmd.ArgsLenAtDash(); pos >= 0 {
		// passthroughArgs := args[pos+1:]
		args = args[0:pos]
	}
	if err := cmd.ParseFlags(args); err != nil {
		return err
	}
	var writer io.Writer
	output, _ := cmd.Flags().GetString("output")
	if output == "-" {
		writer = os.Stdout
	} else if output == "" {
		writer = io.Discard
	} else {
		if w, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			return err
		} else {
			writer = w
		}
	}

	return pipeviz.Graph(writer, args)
}

func RunHandler(cmd *cobra.Command, args []string) error {
	if pos := cmd.ArgsLenAtDash(); pos >= 0 {
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
