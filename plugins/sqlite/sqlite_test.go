package sqlite_test

import (
	"os"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/psutil"
	_ "github.com/OutOfBedlam/tine/plugins/sqlite"
)

func TestSqlite(t *testing.T) {
	seq := int64(0)
	engine.Now = func() time.Time { seq++; return time.Unix(1721954797+seq, 0) }
	output, err := engine.New(engine.WithConfig(output))
	if err != nil {
		t.Fatal(err)
	}
	output.Start()
	time.Sleep(500 * time.Millisecond)

	input, err := engine.New(engine.WithConfig(input), engine.WithWriter(os.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	err = input.Run()
	if err != nil {
		t.Fatal(err)
	}
}

const input = `
[log]
  path = "-"
  level = "warn"
  no_color = true
[[inlets.sqlite]]
  interval = "1s"
  count = 2
  path = "file::memdb?mode=memory&cache=shared"
  actions = [
	["SELECT time, name, value FROM test ORDER BY time"],
  ]
[[outlets.file]]
  path = "-"
  format = "json"
`

const output = `
[log]
  path = "-"
  level = "warn"
  no_color = true
[[inlets.cpu]]
  percpu = false
  interval = "1s"
  count = 2
[[flows.flatten]]
[[outlets.sqlite]]
	path = "file::memdb?mode=memory&cache=shared"
	inits = [
		"CREATE TABLE test (time INTEGER, name TEXT, value REAL, UNIQUE(time, name))",
	]
	actions = [
		["INSERT INTO test (time, name, value) VALUES (?, ?, ?)", "_ts", "name", "value"],
	]
#[[outlets.file]]
#  path = "-"
#  format = "json"
`
