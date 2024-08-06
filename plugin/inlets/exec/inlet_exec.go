package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "exec",
		Factory: ExecInlet,
	})
}

func ExecInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	ret := &execInlet{ctx: ctx}
	ret.commands = conf.GetStringSlice("commands", []string{})
	ret.environments = conf.GetStringSlice("environments", []string{})
	ret.interval = conf.GetDuration("interval", 0)
	ret.namePrefix = conf.GetString("prefix", "")
	ret.ignoreError = conf.GetBool("ignore_error", false)
	ret.timeout = conf.GetDuration("timeout", 0)
	ret.separator = []byte(conf.GetString("separator", ""))
	ret.runCountLimit = int64(conf.GetInt("count", 0))
	ret.trimSpace = conf.GetBool("trim_space", false)
	return ret
}

type execInlet struct {
	ctx *engine.Context

	commands      []string
	environments  []string
	namePrefix    string
	separator     []byte
	interval      time.Duration
	ignoreError   bool
	timeout       time.Duration
	runCountLimit int64
	trimSpace     bool

	cmd      *exec.Cmd
	runCount int64
}

var _ = engine.Inlet((*execInlet)(nil))

func (ei *execInlet) Open() error {
	if len(ei.commands) == 0 {
		return fmt.Errorf("exec commands not specified")
	}
	return nil
}

func (ei *execInlet) Close() error {
	if ei.cmd != nil {
		ei.cmd.Process.Kill()
		ei.cmd.Wait()
		ei.cmd = nil
	}
	return nil
}

func (ei *execInlet) Interval() time.Duration {
	return ei.interval
}

func (ei *execInlet) Process(next engine.InletNextFunc) {
	if ei.runCountLimit > 0 && atomic.LoadInt64(&ei.runCount) >= ei.runCountLimit {
		next(nil, io.EOF)
		return
	}

	ei.cmd = exec.Command(ei.commands[0], ei.commands[1:]...)
	ei.cmd.Env = append(os.Environ(), ei.environments...)
	stdoutWriter := &execWriter{name: ei.namePrefix + "stdout", next: next, separator: ei.separator, trimSpace: ei.trimSpace}
	stderrWriter := &execWriter{name: ei.namePrefix + "stderr", next: next, separator: ei.separator, trimSpace: ei.trimSpace}
	ei.cmd.Stdout = stdoutWriter
	ei.cmd.Stderr = stderrWriter

	var resultErr error
	defer func() {
		stdoutWriter.Close()
		stderrWriter.Close()
		if resultErr != nil {
			next(nil, resultErr)
		}
	}()

	runCount := atomic.AddInt64(&ei.runCount, 1)
	ei.ctx.LogDebug("exec run", "cmd", strings.Join(ei.commands, " "), "interval", ei.interval, "count", runCount, "countLimit", ei.runCountLimit)

	if err := ei.cmd.Start(); err != nil {
		ei.ctx.LogError("exec start error", "err", err)
		next(nil, err)
		return
	}

	doneC := make(chan struct{})
	go func() {
		if err := ei.cmd.Wait(); err != nil {
			resultErr = fmt.Errorf("exec [%s] error: %s", ei.commands[0], err)
		} else {
			state := ei.cmd.ProcessState
			if !ei.ignoreError && state.ExitCode() != 0 {
				resultErr = fmt.Errorf("exec [%s] exit: %d", ei.commands[0], state.ExitCode())
			} else if ei.runCountLimit > 0 && runCount > ei.runCountLimit {
				resultErr = io.EOF
			}
		}
		close(doneC)
	}()

	if ei.timeout > 0 {
		select {
		case <-doneC:
		case <-time.After(ei.timeout):
			ei.cmd.Process.Kill()
			ei.ctx.LogDebug("exec timeout", "cmd", strings.Join(ei.commands, " "))
			next(nil, fmt.Errorf("exec [%s] timeout", ei.commands[0]))
		}
	} else {
		<-doneC
	}
}

type execWriter struct {
	name      string
	next      engine.InletNextFunc
	separator []byte
	trimSpace bool
	buff      []byte
	offset    int
}

func (ew *execWriter) Write(p []byte) (n int, err error) {
	ret := len(p)
	if ew.buff == nil {
		ew.buff = make([]byte, 4096)
		ew.offset = 0
	}
	if len(ew.separator) == 0 {
		if ew.trimSpace {
			p = bytes.TrimSpace(p)
		}
		ew.next([]engine.Record{
			engine.NewRecord(engine.NewField(ew.name, string(p))),
		}, nil)
	} else {
		lines := bytes.Split(p, ew.separator)
		if len(lines) > 1 {
			ew.offset = copy(ew.buff, lines[len(lines)-1])
			lines = lines[0 : len(lines)-1]
		}

		ret := []engine.Record{}
		for _, line := range lines {
			ret = append(ret, engine.NewRecord(
				engine.NewField(ew.name, string(line)),
			))
		}
		ew.next(ret, nil)
	}
	return ret, nil
}

func (ew *execWriter) Close() error {
	if ew.buff != nil && ew.offset > 0 {
		ew.next([]engine.Record{
			engine.NewRecord(engine.NewField(ew.name, string(ew.buff[:ew.offset]))),
		}, nil)
	}
	return nil
}
