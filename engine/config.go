package engine

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/OutOfBedlam/tine/util"
)

type PipelineConfig struct {
	Name     string
	Log      util.LogConfig
	Defaults Config
	Inlets   []InletConfig
	Outlets  []OutletConfig
	Flows    []FlowConfig
}

var topLevelKeys = []string{"inlets", "outlets", "flows", "defaults", "log", "name"}

func LoadConfig(content string, cfg *PipelineConfig) error {
	vc := NewConfig()
	meta, err := toml.Decode(content, &vc)
	if err != nil {
		return err
	}
	cfg.Name = vc.GetString("name", cfg.Name)
	if lc := vc.GetConfig("log", nil); lc != nil {
		cfg.Log.Filename = lc.GetString("filename", cfg.Log.Filename)
		cfg.Log.Level = lc.GetString("level", cfg.Log.Level)
		cfg.Log.MaxSize = lc.GetInt("max_size", cfg.Log.MaxSize)
		cfg.Log.MaxAge = lc.GetInt("max_age", cfg.Log.MaxAge)
		cfg.Log.MaxBackups = lc.GetInt("max_backups", cfg.Log.MaxBackups)
		cfg.Log.Compress = lc.GetBool("compress", cfg.Log.Compress)
		cfg.Log.Chown = lc.GetString("chown", cfg.Log.Chown)
	}
	cfg.Defaults = vc.GetConfig("defaults", cfg.Defaults)

	sameKeys := map[string]int{}
	for _, keys := range meta.Keys() {
		if len(keys) == 2 && (keys[0] == "inlets" || keys[0] == "outlets" || keys[0] == "flows") {
			kind := keys[0]
			pluginName := keys[1]
			inletCfg := vc.GetConfig(kind, nil)
			paramsArr := inletCfg.GetConfigSlice(pluginName, nil)
			var params Config
			if len(paramsArr) == 0 {
				params = Config{}
			} else if len(paramsArr) > 1 {
				idx := sameKeys[pluginName]
				params = paramsArr[idx]
				sameKeys[pluginName]++
			} else {
				params = paramsArr[0]
			}
			flowCfgs := params.GetConfig("flows", nil)
			params.Unset("flows")
			flows := []FlowConfig{}
			if kind != "flows" {
				flowsInOrder := []string{}
				for _, keys := range meta.Keys() {
					if len(keys) == 4 && keys[0] == kind && keys[1] == pluginName && keys[2] == "flows" {
						flowsInOrder = append(flowsInOrder, keys[3])
					}
				}
				for _, flowName := range flowsInOrder {
					flowParam := flowCfgs.GetConfig(flowName, nil)
					flows = append(flows, FlowConfig{
						Plugin: flowName,
						Params: flowParam,
					})
				}
			}
			if kind == "inlets" {
				cfg.Inlets = append(cfg.Inlets, InletConfig{
					Plugin: pluginName,
					Params: params,
					Flows:  flows,
				})
			} else if kind == "outlets" {
				cfg.Outlets = append(cfg.Outlets, OutletConfig{
					Plugin: pluginName,
					Params: params,
					//Flows:  flows,
				})
			} else if kind == "flows" {
				cfg.Flows = append(cfg.Flows, FlowConfig{
					Plugin: pluginName,
					Params: params,
				})
			}
		} else if len(keys) > 0 {
			if !slices.Contains(topLevelKeys, keys[0]) {
				return fmt.Errorf("unexpected keys %s", keys)
			}
		} else {
			return fmt.Errorf("unexpected key %s", keys)
		}
	}
	return nil
}

type InletConfig struct {
	Plugin string
	Params Config
	Flows  []FlowConfig
}

type OutletConfig struct {
	Plugin string
	Params Config
	//Flows  []FlowConfig
}

type FlowConfig struct {
	Plugin string
	Params Config
}

type Config map[string]any

func NewConfig() Config {
	return make(map[string]any)
}

func (c Config) Set(key string, val any) Config {
	c[key] = val
	return c
}

func (c Config) Unset(key string) Config {
	delete(c, key)
	return c
}

func (c Config) GetConfig(key string, defaultVal Config) Config {
	if v, ok := c[key]; ok {
		switch v := v.(type) {
		case map[string]any:
			return v
		case []map[string]any:
			if len(v) > 0 {
				return v[0]
			}
		}
	}
	return defaultVal
}

func (c Config) GetConfigSlice(key string, defaultVal []Config) []Config {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case map[string]any:
			return []Config{val}
		case []map[string]any:
			result := make([]Config, 0, len(val))
			for _, x := range val {
				result = append(result, x)
			}
			return result
		case []any:
			result := make([]Config, 0, len(val))
			for _, x := range val {
				if xv, ok := x.(map[string]any); ok {
					result = append(result, xv)
				}
			}
			return result
		}
	}
	return defaultVal
}

func (c Config) GetBool(key string, defaultVal bool) bool {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			if result, err := strconv.ParseBool(val); err == nil {
				return result
			}
		case int:
			return val != 0
		}
	}
	return defaultVal
}

func (c Config) GetDuration(key string, defaultVal time.Duration) time.Duration {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case time.Duration:
			return val
		case string:
			if result, err := time.ParseDuration(val); err == nil {
				return result
			}
		}
	}
	return defaultVal
}

func (c Config) GetString(key string, defaultVal string) string {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case int:
			return strconv.Itoa(val)
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(val)
		}
	}
	return defaultVal
}

func (c Config) GetStringSlice(key string, defaultVal []string) []string {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case []string:
			return val
		case []any:
			result := make([]string, 0, len(val))
			for _, x := range val {
				if xv, ok := x.(string); ok {
					result = append(result, xv)
				}
			}
			return result
		case string:
			return []string{val}
		}
	}
	return defaultVal
}

func (c Config) GetInt(key string, defaultVal int) int {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case uint:
			return int(val)
		case int64:
			return int(val)
		case uint64:
			return int(val)
		case string:
			if result, err := strconv.ParseInt(val, 10, 32); err == nil {
				return int(result)
			}
		}
	}
	return defaultVal
}

func (c Config) GetInt64(key string, defaultVal int64) int64 {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case int:
			return int64(val)
		case uint:
			return int64(val)
		case int32:
			return int64(val)
		case uint32:
			return int64(val)
		case int64:
			return val
		case uint64:
			return int64(val)
		case string:
			if result, err := strconv.ParseInt(val, 10, 64); err == nil {
				return result
			}
		}
	}
	return defaultVal
}

func (c Config) GetFloat(key string, defaultVal float64) float64 {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case int:
			return float64(val)
		case uint:
			return float64(val)
		case int64:
			return float64(val)
		case uint64:
			return float64(val)
		case float32:
			return float64(val)
		case float64:
			return val
		case string:
			if result, err := strconv.ParseFloat(val, 64); err == nil {
				return result
			}
		}
	}
	return defaultVal
}

func (c Config) GetIntSlice(key string, defaultVal []int) []int {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case []int:
			return val
		case int:
			return []int{val}
		case int64:
			return []int{int(val)}
		case []any:
			result := make([]int, 0, len(val))
			for _, x := range val {
				switch xv := x.(type) {
				case int64:
					result = append(result, int(xv))
				case int:
					result = append(result, xv)
				}
			}
			return result
		}
	}
	return defaultVal
}

func (c Config) GetUint32(key string, defaultVal uint32) uint32 {
	if v, ok := c[key]; ok {
		switch val := v.(type) {
		case uint32:
			return val
		case int:
			return uint32(val)
		case string:
			if result, err := strconv.ParseUint(val, 10, 64); err == nil {
				return uint32(result)
			}
		}

	}
	return defaultVal
}
