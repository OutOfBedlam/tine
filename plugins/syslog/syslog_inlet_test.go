package syslog_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/syslog"
	"github.com/stretchr/testify/require"
)

func TestSyslogInlet(t *testing.T) {
	tests := []struct {
		pipeFile   string
		inputFile  string
		expectFile string
		skip       func() bool
		check      func(string) bool
	}{
		{
			pipeFile:   "./testdata/syslog1.toml",
			inputFile:  "./testdata/syslog1_in.txt",
			expectFile: "./testdata/syslog1_out.txt",
		},
	}

	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	engine.Now = func() time.Time { return time.Unix(1724243010, 0) }

	for _, tt := range tests {
		if tt.skip != nil && tt.skip() {
			continue
		}
		runTest(t, tt.pipeFile, tt.inputFile, tt.expectFile, tt.check)
	}
}

func runTest(t *testing.T, pipeFile, inputFile, expectFile string, check func(string) bool) {
	t.Helper()

	p, err := engine.New(engine.WithConfigFile(pipeFile))
	if err != nil {
		t.Errorf("failed to parse pipeline %s: %v", pipeFile, err)
		t.Fail()
		return
	}

	r, w, _ := os.Pipe()
	os.Stdout = w

	wgPipeline := sync.WaitGroup{}
	wgPipeline.Add(1)
	go func() {
		if err := p.Run(); err != nil {
			t.Errorf("failed to run pipeline %s: %v", pipeFile, err)
			t.Fail()
		}
		wgPipeline.Done()
	}()

	b := new(bytes.Buffer)
	wgCopy := sync.WaitGroup{}
	wgCopy.Add(1)
	go func() {
		io.Copy(b, r)
		wgCopy.Done()
	}()

	wgSender := sync.WaitGroup{}
	wgSender.Add(1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("SEND...")
		syslog_sender(inputFile)
		time.Sleep(100 * time.Millisecond)
		fmt.Println("SEND...DONE")
		wgSender.Done()
	}()
	wgSender.Wait()
	p.Stop()
	wgPipeline.Wait()
	w.Close()
	wgCopy.Wait()

	expectBin, err := os.ReadFile(expectFile)
	if err != nil {
		t.Fatal(err)
	}
	buff := []string{}
	for _, line := range strings.Split(string(expectBin), "\n") {
		buff = append(buff, strings.TrimSpace(line))
	}
	expect := strings.Join(buff, "\n")

	buff = []string{}
	for _, line := range strings.Split(b.String(), "\n") {
		buff = append(buff, strings.TrimSpace(line))
	}
	result := strings.Join(buff, "\n")

	if check != nil {
		if !check(result) {
			t.Errorf("unexpected output for %s", pipeFile)
			t.Fail()
		}
		return
	}
	require.Equal(t, expect, result, pipeFile)
}

func syslog_sender(inputFile string) {
	// Connect to the syslog server over UDP
	writer, err := net.Dial("udp", "127.0.0.1:5516")
	if err != nil {
		slog.Warn("Failed to connect to syslog server:", "error", err)
	}
	defer writer.Close()

	// Send a log message
	content, _ := os.ReadFile(inputFile)
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		writer.Write(line)
	}
}
