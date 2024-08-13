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

type ExecDriver struct {
	ctx          *engine.Context
	commands     []string
	environments []string
	ignoreError  bool
	timeout      time.Duration
	// inlet options
	interval      time.Duration
	runCountLimit int64
	// flow options
	parallelism int

	// output options
	namePrefix string
	separator  []byte
	trimSpace  bool

	cmd      *exec.Cmd
	runCount int64
}

func New(ctx *engine.Context) *ExecDriver {
	return &ExecDriver{ctx: ctx}
}

func (ed *ExecDriver) Open() error {
	conf := ed.ctx.Config()
	ed.commands = conf.GetStringSlice("commands", []string{})
	ed.environments = conf.GetStringSlice("environments", []string{})
	ed.namePrefix = conf.GetString("prefix", "")
	ed.ignoreError = conf.GetBool("ignore_error", false)
	ed.timeout = conf.GetDuration("timeout", 0)
	ed.separator = []byte(conf.GetString("separator", ""))
	ed.trimSpace = conf.GetBool("trim_space", false)
	// inlet options
	ed.runCountLimit = int64(conf.GetInt("count", 0))
	ed.interval = conf.GetDuration("interval", 0)
	// flow options
	ed.parallelism = conf.GetInt("parallelism", 1)

	if len(ed.commands) == 0 {
		return fmt.Errorf("exec commands not specified")
	}
	return nil
}

func (ed *ExecDriver) Close() error {
	if ed.cmd != nil {
		ed.cmd.Process.Kill()
		ed.cmd.Wait()
		ed.cmd = nil
	}
	return nil
}

func (ed *ExecDriver) Interval() time.Duration {
	return ed.interval
}

func (ed *ExecDriver) Parallelism() int {
	return ed.parallelism
}

func (ed *ExecDriver) Process(inputRecord engine.Record, next func([]engine.Record, error)) {
	if ed.runCountLimit > 0 && atomic.LoadInt64(&ed.runCount) >= ed.runCountLimit {
		next(nil, io.EOF)
		return
	}

	var tags engine.Tags
	var vals []string
	if inputRecord != nil {
		tags = inputRecord.Tags()
		for _, field := range inputRecord.Fields() {
			v, _ := field.Value.String()
			ln := fmt.Sprintf("FIELD_%s=%s", strings.ToUpper(field.Name), v)
			vals = append(vals, ln)
		}
		for k, val := range inputRecord.Tags() {
			v, _ := val.String()
			ln := fmt.Sprintf("TAG_%s=%s", strings.ToUpper(k), v)
			vals = append(vals, ln)
		}
	}

	ed.cmd = exec.Command(ed.commands[0], ed.commands[1:]...)
	ed.cmd.Env = append(os.Environ(), ed.environments...)
	ed.cmd.Env = append(ed.cmd.Env, vals...)

	stdoutWriter := &execWriter{name: ed.namePrefix + "stdout", tags: tags, next: next, separator: ed.separator, trimSpace: ed.trimSpace}
	stderrWriter := &execWriter{name: ed.namePrefix + "stderr", tags: tags, next: next, separator: ed.separator, trimSpace: ed.trimSpace}
	ed.cmd.Stdout = stdoutWriter
	ed.cmd.Stderr = stderrWriter

	var resultErr error
	defer func() {
		stdoutWriter.Close()
		stderrWriter.Close()
		if resultErr != nil {
			next(nil, resultErr)
		}
	}()

	runCount := atomic.AddInt64(&ed.runCount, 1)
	ed.ctx.LogDebug("exec run", "cmd", strings.Join(ed.commands, " "), "interval", ed.interval, "count", runCount, "countLimit", ed.runCountLimit)

	if err := ed.cmd.Start(); err != nil {
		ed.ctx.LogError("exec start error", "err", err)
		next(nil, err)
		return
	}

	doneC := make(chan struct{})
	go func() {
		if err := ed.cmd.Wait(); err != nil {
			resultErr = fmt.Errorf("exec [%s] error: %s", ed.commands[0], err)
		} else {
			state := ed.cmd.ProcessState
			if !ed.ignoreError && state.ExitCode() != 0 {
				resultErr = fmt.Errorf("exec [%s] exit: %d", ed.commands[0], state.ExitCode())
			} else if ed.runCountLimit > 0 && runCount > ed.runCountLimit {
				resultErr = io.EOF
			}
		}
		close(doneC)
	}()

	if ed.timeout > 0 {
		select {
		case <-doneC:
		case <-time.After(ed.timeout):
			ed.cmd.Process.Kill()
			ed.ctx.LogDebug("exec timeout", "cmd", strings.Join(ed.commands, " "))
			next(nil, fmt.Errorf("exec [%s] timeout", ed.commands[0]))
		}
	} else {
		<-doneC
	}
}

type execWriter struct {
	name      string
	tags      engine.Tags
	next      func([]engine.Record, error)
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

	var outputRecords []engine.Record
	if len(ew.separator) == 0 {
		if ew.trimSpace {
			p = bytes.TrimSpace(p)
		}
		rec := engine.NewRecord(engine.NewField(ew.name, string(p)))
		for k, v := range ew.tags {
			rec.Tags()[k] = v
		}
		outputRecords = []engine.Record{rec}
	} else {
		lines := bytes.Split(p, ew.separator)
		if len(lines) > 1 {
			ew.offset = copy(ew.buff, lines[len(lines)-1])
			lines = lines[0 : len(lines)-1]
		}

		for _, line := range lines {
			rec := engine.NewRecord(engine.NewField(ew.name, string(line)))
			for k, v := range ew.tags {
				rec.Tags()[k] = v
			}
			outputRecords = append(outputRecords, rec)
		}
	}
	ew.next(outputRecords, nil)
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
