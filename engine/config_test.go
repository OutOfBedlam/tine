package engine

import (
	"math"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/util"
	"github.com/stretchr/testify/require"
)

var testDefaultLogConfig = util.LogConfig{
	Path:       "-",
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

func TestConfigArrayItem(t *testing.T) {
	content := `
	[defaults]
		interval = "1s"
	[[inlets.cpu]]
		sval = [ "s1", "s2", "s3" ]
		ival = [ 1, 2, 3]
	`
	actual := PipelineConfig{}

	if err := LoadConfig(content, &actual); err != nil {
		t.Log("ERROR", err.Error())
		t.Fail()
	}
	require.Equal(t, []any{"s1", "s2", "s3"}, actual.Inlets[0].Params["sval"])
	require.Equal(t, []any{int64(1), int64(2), int64(3)}, actual.Inlets[0].Params["ival"])
	conf := makeConfig(actual.Inlets[0].Params, actual.Defaults)
	require.Equal(t, time.Second, conf.GetDuration("interval", 0))
}

func TestConfigValues(t *testing.T) {
	c := Config{
		"str":    "string",
		"int":    1,
		"uint":   uint(2),
		"int64":  int64(3),
		"uint64": uint64(4),
		"sint":   "5",
		"int32":  int32(6),
		"uint32": uint32(7),
		"flt32":  float32(3.14),
		"sflt":   "3.14",
		"flt64":  float64(3.14),
		"bol":    true,
		"sbol":   "true",
		"arr":    []any{"a", "b", "c"},
		"map": map[string]any{
			"str": "string",
			"int": 1,
			"flt": 3.14,
			"bol": true,
			"arr": []any{"a", "b", "c"},
		},
		"smallMap": map[string]any{
			"str": "small",
		},
	}

	c.Set("str", "new string")
	require.Equal(t, "new string", c.GetString("str", ""))
	require.Equal(t, []Config{{"str": "small"}}, c.GetConfigSlice("smallMap", nil))
	require.Equal(t, []string{"a", "b", "c"}, c.GetStringSlice("arr", nil))
	// GetBool
	require.Equal(t, true, c.GetBool("bol", false))
	require.Equal(t, true, c.GetBool("int", false))
	require.Equal(t, true, c.GetBool("sbol", false))
	require.Equal(t, true, c.GetBool("flt32", false))
	require.Equal(t, true, c.GetBool("flt64", false))
	// GetString
	require.Equal(t, "new string", c.GetString("str", ""))
	require.Equal(t, "1", c.GetString("int", ""))
	require.Equal(t, "2", c.GetString("uint", ""))
	require.Equal(t, "3", c.GetString("int64", ""))
	require.Equal(t, "4", c.GetString("uint64", ""))
	require.Equal(t, "5", c.GetString("sint", ""))
	require.Equal(t, "6", c.GetString("int32", ""))
	require.Equal(t, "7", c.GetString("uint32", ""))
	require.Equal(t, "3.14", c.GetString("flt32", ""))
	require.Equal(t, "3.14", c.GetString("flt64", ""))
	require.Equal(t, "true", c.GetString("bol", ""))
	// GetInt
	require.Equal(t, int(1), c.GetInt("int", 0))
	require.Equal(t, int(2), c.GetInt("uint", 0))
	require.Equal(t, int(3), c.GetInt("int64", 0))
	require.Equal(t, int(4), c.GetInt("uint64", 0))
	require.Equal(t, int(5), c.GetInt("sint", 0))
	require.Equal(t, int(6), c.GetInt("int32", 0))
	require.Equal(t, int(7), c.GetInt("uint32", 0))
	// GetInt
	require.Equal(t, uint32(1), c.GetUint32("int", 0))
	require.Equal(t, uint32(2), c.GetUint32("uint", 0))
	require.Equal(t, uint32(3), c.GetUint32("int64", 0))
	require.Equal(t, uint32(4), c.GetUint32("uint64", 0))
	require.Equal(t, uint32(5), c.GetUint32("sint", 0))
	require.Equal(t, uint32(6), c.GetUint32("int32", 0))
	require.Equal(t, uint32(7), c.GetUint32("uint32", 0))
	// GetInt64
	require.Equal(t, int64(1), c.GetInt64("int", 0))
	require.Equal(t, int64(2), c.GetInt64("uint", 0))
	require.Equal(t, int64(3), c.GetInt64("int64", 0))
	require.Equal(t, int64(4), c.GetInt64("uint64", 0))
	require.Equal(t, int64(5), c.GetInt64("sint", 0))
	require.Equal(t, int64(6), c.GetInt64("int32", 0))
	require.Equal(t, int64(7), c.GetInt64("uint32", 0))
	// GetFloat
	require.Equal(t, 3.14, c.GetFloat("flt64", 0.0))
	require.Equal(t, float64(3.14), c.GetFloat("flt64", 0.0))
	require.Equal(t, float64(3.14), math.Round(c.GetFloat("flt32", 0.0)*100)/100)
	require.Equal(t, float64(3.14), c.GetFloat("sflt", 0.0))
	require.Equal(t, float64(1), c.GetFloat("int", 0))
	require.Equal(t, float64(2), c.GetFloat("uint", 0))
	require.Equal(t, float64(3), c.GetFloat("int64", 0))
	require.Equal(t, float64(4), c.GetFloat("uint64", 0))
	require.Equal(t, float64(5), c.GetFloat("sint", 0))
	require.Equal(t, float64(6), c.GetFloat("int32", 0))
	require.Equal(t, float64(7), c.GetFloat("uint32", 0))
	// GetValue
	require.Equal(t, 3.14, c.GetValue("flt64").raw)
	require.Equal(t, 3.14, math.Round(c.GetValue("flt32").raw.(float64)*100)/100)
	require.Equal(t, "3.14", c.GetValue("sflt").raw)
	require.Equal(t, true, c.GetValue("bol").raw)
	require.Equal(t, int64(1), c.GetValue("int").raw)
	require.Equal(t, uint64(2), c.GetValue("uint").raw)
	require.Equal(t, int64(3), c.GetValue("int64").raw)
	require.Equal(t, uint64(4), c.GetValue("uint64").raw)
}
