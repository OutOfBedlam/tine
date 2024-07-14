package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

type execPull struct {
	ctx *engine.Context

	interval     time.Duration
	commands     []string
	environments []string
	namePrefix   string
	ignoreError  bool
	timeout      time.Duration
	countLimit   int64
	trimSpace    bool

	cmd       *exec.Cmd
	runtcount int64
}

var _ = engine.PullInlet((*execPull)(nil))

func (ei *execPull) Open() error {
	if len(ei.commands) == 0 {
		return fmt.Errorf("exec commands not specified")
	}
	return nil
}

func (ei *execPull) Close() error { return nil }

// Implements engine.PullInlet, for batch mode
func (ei *execPull) Interval() time.Duration {
	return ei.interval
}

// Implements engine.PullInlet, for batch mode
// return io.EOF when countLimit reached
func (ei *execPull) Pull() ([]engine.Record, error) {
	if ei.cmd != nil {
		ei.ctx.LogDebug("exec is already running", "cmd", strings.Join(ei.commands, " "))
		return nil, nil
	}
	runCount := atomic.AddInt64(&ei.runtcount, 1)
	ei.ctx.LogDebug("exec run", "cmd", strings.Join(ei.commands, " "), "count", runCount, "countLimit", ei.countLimit)

	ei.cmd = exec.Command(ei.commands[0], ei.commands[1:]...)
	defer func() {
		ei.cmd = nil
	}()
	ei.cmd.Env = append(os.Environ(), ei.environments...)

	stdout, err := ei.cmd.StdoutPipe()
	if err != nil {
		ei.ctx.LogError("exec stdout pipe error", "err", err)
		return nil, err
	}

	stderr, err := ei.cmd.StderrPipe()
	if err != nil {
		ei.ctx.LogError("exec stderr pipe error", "err", err)
		return nil, err
	}

	if err := ei.cmd.Start(); err != nil {
		ei.ctx.LogError("exec start error", "err", err)
		return nil, err
	}

	done := make(chan *os.ProcessState, 1)
	go func() {
		state, _ := ei.cmd.Process.Wait()
		done <- state
	}()

	select {
	case <-time.After(ei.timeout):
		ei.cmd.Process.Kill()
		outbytes, err := io.ReadAll(stdout)
		if err != nil {
			return nil, err
		}
		errbytes, err := io.ReadAll(stderr)
		if err != nil {
			return nil, err
		}
		strout := string(outbytes)
		strerr := string(errbytes)
		if ei.trimSpace {
			strout = strings.TrimSpace(strout)
			strerr = strings.TrimSpace(strerr)
		}
		ei.ctx.LogDebug("exec timeout", "cmd", strings.Join(ei.commands, " "), "stdout", strout, "stderr", strerr)
		return nil, fmt.Errorf("exec [%s] timeout", ei.commands[0])
	case state := <-done:
		if !ei.ignoreError && state.ExitCode() != 0 {
			return nil, fmt.Errorf("exec [%s] error: %s", ei.commands[0], state.String())
		}
		outbytes, err := io.ReadAll(stdout)
		if err != nil {
			return nil, err
		}
		errbytes, err := io.ReadAll(stderr)
		if err != nil {
			return nil, err
		}
		exitCode := state.ExitCode()

		strout := string(outbytes)
		strerr := string(errbytes)
		if ei.trimSpace {
			strout = strings.TrimSpace(strout)
			strerr = strings.TrimSpace(strerr)
		}
		ret := []engine.Record{
			engine.NewRecord(
				engine.NewIntField(ei.namePrefix+"exit_code", int64(exitCode)),
				engine.NewStringField(ei.namePrefix+"stdout", strout),
				engine.NewStringField(ei.namePrefix+"stderr", strerr),
			),
		}
		if ei.countLimit > 0 && runCount >= ei.countLimit {
			ei.ctx.LogDebug("exec count limit reached", "count", runCount, "limit", ei.countLimit)
			return ret, io.EOF
		}
		return ret, nil
	}
}
