package engine

import (
	"testing"

	"github.com/OutOfBedlam/tine/util"
	"github.com/stretchr/testify/require"
)

var testDefaultLogConfig = util.LogConfig{
	Filename:   "-",
	Level:      util.LOG_LEVEL_INFO,
	MaxSize:    100,
	MaxAge:     7,
	MaxBackups: 10,
	Compress:   false,
	Chown:      "",
}

var testDefaultConfig = map[string]any{
	"interval": "10s",
}

func TestConfig(t *testing.T) {
	tests := []struct {
		content string
		expect  PipelineConfig
	}{
		{
			content: `
			name = "test"
			[[inlets.cpu]]
				percpu = true
			[[inlets.load]]
			[[outlets.file]]
				path = "test.log"
			[[outlets.file]]
				path = "test2.log"
			[[outlets.file]]
				path = "test3.log"
			`,

			expect: PipelineConfig{
				Name:     "test",
				Log:      testDefaultLogConfig,
				Defaults: testDefaultConfig,
				Inlets: map[string][]Config{
					"cpu": {
						{"percpu": true},
					},
					"load": {{}},
				},
				Outlets: map[string][]Config{
					"file": {
						{"path": "test.log"},
						{"path": "test2.log"},
						{"path": "test3.log"},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		actual, err := New(WithConfig(tc.content))
		if err != nil {
			t.Log("ERROR", err.Error())
			t.Fail()
		}
		require.Equal(t, tc.expect, actual.PipelineConfig)
	}
}
