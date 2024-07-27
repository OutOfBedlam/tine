package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/OutOfBedlam/tine/engine"
)

type execPush struct {
	ctx          *engine.Context
	commands     []string
	environments []string
	namePrefix   string
	separator    []byte
	trimSpace    bool

	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

var _ = engine.PushInlet((*execPush)(nil))

func (ep *execPush) Open() error {
	if len(ep.commands) == 0 {
		return fmt.Errorf("exec commands not specified")
	}
	ep.cmd = exec.Command(ep.commands[0], ep.commands[1:]...)
	ep.cmd.Env = append(os.Environ(), ep.environments...)
	if stdout, err := ep.cmd.StdoutPipe(); err != nil {
		return err
	} else {
		ep.stdout = stdout
	}
	if stderr, err := ep.cmd.StderrPipe(); err != nil {
		return err
	} else {
		ep.stderr = stderr
	}
	return ep.cmd.Start()
}

func (ep *execPush) Close() error {
	if ep.cmd != nil {
		ep.cmd.Process.Kill()
		ep.cmd.Wait()
		ep.cmd = nil
	}
	return nil
}

func (ep *execPush) Push(fn func([]engine.Record, error)) {
	if ep.cmd == nil {
		fn(nil, fmt.Errorf("exec command not started"))
		return
	}
	buff := make([]byte, 4096)
	offset := 0
	for {
		n, err := ep.stdout.Read(buff[offset:])
		if err != nil {
			fn(nil, err)
			break
		}
		if len(ep.separator) == 0 {
			result := buff[:n]
			if ep.trimSpace {
				result = bytes.TrimSpace(result)
			}
			fn([]engine.Record{engine.NewRecord(
				engine.NewField(ep.namePrefix+"stdout", string(result)),
			)}, nil)
		} else {
			lines := bytes.Split(buff[:n], ep.separator)
			if len(lines) > 1 {
				offset = copy(buff, lines[len(lines)-1])
				lines = lines[0 : len(lines)-1]
			}

			ret := []engine.Record{}
			for _, line := range lines {
				ret = append(ret, engine.NewRecord(
					engine.NewField(ep.namePrefix+"stdout", string(line)),
				))
			}
			fn(ret, nil)
		}
	}
}
