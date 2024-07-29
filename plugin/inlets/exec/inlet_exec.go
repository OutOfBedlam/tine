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
	stdout   io.ReadCloser
	stderr   io.ReadCloser
	runcount int64
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
	if ei.runCountLimit > 0 && atomic.LoadInt64(&ei.runcount) >= ei.runCountLimit {
		next(nil, io.EOF)
		return
	}

	ei.cmd = exec.Command(ei.commands[0], ei.commands[1:]...)
	ei.cmd.Env = append(os.Environ(), ei.environments...)
	if stdout, err := ei.cmd.StdoutPipe(); err != nil {
		next(nil, err)
		return
	} else {
		ei.stdout = stdout
	}
	if stderr, err := ei.cmd.StderrPipe(); err != nil {
		next(nil, err)
		return
	} else {
		ei.stderr = stderr
	}

	runCount := atomic.AddInt64(&ei.runcount, 1)
	ei.ctx.LogDebug("exec run", "cmd", strings.Join(ei.commands, " "), "interval", ei.interval, "count", runCount, "countLimit", ei.runCountLimit)

	if err := ei.cmd.Start(); err != nil {
		ei.ctx.LogError("exec start error", "err", err)
		next(nil, err)
		return
	}

	go func() {
		buff := make([]byte, 4096)
		offset := 0
		for {
			var yieldErr error
			var yieldRecs []engine.Record

			n, err := ei.stdout.Read(buff[offset:])
			if err != nil {
				yieldErr = err
			}
			if n > 0 {
				if len(ei.separator) == 0 {
					result := buff[:n]
					if ei.trimSpace {
						result = bytes.TrimSpace(result)
					}
					yieldRecs = []engine.Record{
						engine.NewRecord(engine.NewField(ei.namePrefix+"stdout", string(result))),
					}
				} else {
					lines := bytes.Split(buff[:n], ei.separator)
					if len(lines) > 1 {
						offset = copy(buff, lines[len(lines)-1])
						lines = lines[0 : len(lines)-1]
					}

					ret := []engine.Record{}
					for _, line := range lines {
						ret = append(ret, engine.NewRecord(
							engine.NewField(ei.namePrefix+"stdout", string(line)),
						))
					}
					yieldRecs = ret
				}
			}
			next(yieldRecs, yieldErr)
			if yieldErr != nil {
				break
			}
		}
	}()

	if ei.timeout > 0 {
		go func() {
			<-time.After(ei.timeout)
			outbytes, err := io.ReadAll(ei.stdout)
			if err != nil {
				next(nil, err)
				return
			}
			errbytes, err := io.ReadAll(ei.stderr)
			if err != nil {
				next(nil, err)
				return
			}
			strout := string(outbytes)
			strerr := string(errbytes)
			if ei.trimSpace {
				strout = strings.TrimSpace(strout)
				strerr = strings.TrimSpace(strerr)
			}
			ei.cmd.Process.Kill()
			ei.ctx.LogDebug("exec timeout", "cmd", strings.Join(ei.commands, " "), "stdout", strout, "stderr", strerr)
			next(nil, fmt.Errorf("exec [%s] timeout", ei.commands[0]))
		}()
	}

	state, err := ei.cmd.Process.Wait()
	if !ei.ignoreError && state.ExitCode() != 0 && err != nil {
		next(nil, fmt.Errorf("exec [%s] error: %s", ei.commands[0], err))
	}
}
