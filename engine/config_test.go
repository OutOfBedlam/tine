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

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		content string
		expect  PipelineConfig
	}{
		{
			content: `
				name = "test"
				[[inlets.cpu]]
					percpu = true
					[[inlets.cpu.flows.f1]]
					[[inlets.cpu.flows.f2]]
					[[inlets.cpu.flows.f3]]
					[[inlets.cpu.flows.f4]]
					[[inlets.cpu.flows.f5]]
				[[inlets.load]]
				[[outlets.file]]
					path = "test.csv"
				[[flows.x1]]
				[[flows.x2]]
				[[flows.x3]]
				[[flows.x4]]
				[[flows.x5]]
			`,
			expect: PipelineConfig{
				Name:     "test",
				Log:      testDefaultLogConfig,
				Defaults: testDefaultConfig,
				Inlets: []InletConfig{
					{
						Plugin: "cpu",
						Params: map[string]any{
							"percpu": true,
						},
						Flows: []FlowConfig{
							{
								Plugin: "f1",
								Params: map[string]any{},
							},
							{
								Plugin: "f2",
								Params: map[string]any{},
							},
							{
								Plugin: "f3",
								Params: map[string]any{},
							},
							{
								Plugin: "f4",
								Params: map[string]any{},
							},
							{
								Plugin: "f5",
								Params: map[string]any{},
							},
						},
					},
					{
						Plugin: "load",
						Params: map[string]any{},
						Flows:  []FlowConfig{},
					},
				},
				Outlets: []OutletConfig{
					{
						Plugin: "file",
						Params: map[string]any{
							"path": "test.csv",
						},
					},
				},
				Flows: []FlowConfig{
					{
						Plugin: "x1",
						Params: map[string]any{},
					},
					{
						Plugin: "x2",
						Params: map[string]any{},
					},
					{
						Plugin: "x3",
						Params: map[string]any{},
					},
					{
						Plugin: "x4",
						Params: map[string]any{},
					},
					{
						Plugin: "x5",
						Params: map[string]any{},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		var actual = PipelineConfig{
			Log:      testDefaultLogConfig,
			Defaults: testDefaultConfig,
		}
		if err := LoadConfig(tc.content, &actual); err != nil {
			t.Log("ERROR", err.Error())
			t.Fail()
		}
		require.Equal(t, tc.expect, actual)
	}
}
