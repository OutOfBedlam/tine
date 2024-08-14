package engine_test

import (
	"strings"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/stretchr/testify/require"
)

func TestContextLog(t *testing.T) {
	dsl := `
	name = 'pipeline-1'
	[log]
		level = "info"
		no_color = true
	[[inlets.file]]
		data = [
			"a,1,1.234,true,2024/08/09 16:01:02", 
			"b,2,2.345,false,2024/08/09 16:03:04", 
			"c,3,3.456,true,2024/08/09 16:05:06",
		]
		format = "csv"
		timeformat = "2006/01/02 15:04:05"
		tz = "UTC"
		fields = ["area","ival","fval","bval","tval"]
		types  = ["string", "int", "float", "bool", "time"]
	[[flows.dump]]
		level = "info"
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	sb := &strings.Builder{}

	// Mock the current time
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }

	p, err := engine.New(engine.WithConfig(dsl), engine.WithWriter(sb), engine.WithLogWriter(sb))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(sb.String(), "\n")
	expects := []string{
		`INF pipeline pipeline-1 start inlets=1 flows=3 outlets=1`,
		`INF pipeline pipeline-1 flow-dump rec=1/3 area=a ival=1 fval=1.234 bval=true tval="2024-08-10 01:01:02"`,
		`INF pipeline pipeline-1 flow-dump rec=2/3 area=b ival=2 fval=2.345 bval=false tval="2024-08-10 01:03:04"`,
		`INF pipeline pipeline-1 flow-dump rec=3/3 area=c ival=3 fval=3.456 bval=true tval="2024-08-10 01:05:06"`,
		`a,1,1.234,true,1723219262`,
		`b,2,2.345,false,1723219384`,
		`c,3,3.456,true,1723219506`,
		`INF pipeline pipeline-1 stop`,
		``,
	}
	require.Equal(t, len(expects), len(lines))
	for i, expect := range expects {
		require.Contains(t, lines[i], expect)
	}
}
