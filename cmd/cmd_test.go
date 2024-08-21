package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		args       []string
		expectFile string
		skip       func() bool
	}{
		{
			args:       []string{"--help"},
			expectFile: "./testdata/help.txt",
		},
		{
			args:       []string{"list"},
			expectFile: "./testdata/list.txt",
			skip: func() bool {
				return engine.GetOutletRegistry("rrd") != nil || runtime.GOOS == "windows"
			},
		},
		{
			args:       []string{"list"},
			expectFile: "./testdata/list-with-rrd.txt",
			skip: func() bool {
				return engine.GetOutletRegistry("rrd") == nil || runtime.GOOS == "windows"
			},
		},
		{
			args:       []string{"graph", "--output", "-", "./testdata/graph1.toml"},
			expectFile: "./testdata/graph1.txt",
		},
		{
			args:       []string{"run", "./testdata/run1.toml"},
			expectFile: "./testdata/run1.txt",
		},
	}

	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	engine.Now = func() time.Time { return time.Unix(1724243010, 0) }

	for _, tt := range tests {
		if tt.skip != nil && tt.skip() {
			continue
		}
		r, w, _ := os.Pipe()
		os.Stdout = w

		output := &bytes.Buffer{}
		outputWg := sync.WaitGroup{}
		outputWg.Add(1)
		go func() {
			io.Copy(output, r)
			outputWg.Done()
		}()

		cmd := NewCmd()
		cmd.SetArgs(tt.args)
		err := cmd.ExecuteContext(context.TODO())
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
		w.Close()
		r.Close()
		outputWg.Wait()

		expectBin, err := os.ReadFile(tt.expectFile)
		if err != nil {
			t.Fatal(err)
		}
		// trim whitespace of the end of the line
		buff := []string{}
		for _, line := range strings.Split(string(expectBin), "\n") {
			buff = append(buff, strings.TrimSpace(line))
		}
		expect := strings.Join(buff, "\n")

		buff = []string{}
		for _, line := range strings.Split(output.String(), "\n") {
			buff = append(buff, strings.TrimSpace(line))
		}
		result := strings.Join(buff, "\n")

		require.Equal(t, expect, result, strings.Join(tt.args, " "))
	}
}
