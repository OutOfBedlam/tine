package engine

import (
	"strconv"
	"time"

	"github.com/OutOfBedlam/tine/util"
)

type PipelineConfig struct {
	Name     string              `toml:"name"`
	Log      util.LogConfig      `toml:"log"`
	Defaults Config              `toml:"defaults"`
	Inlets   map[string][]Config `toml:"inlets,omitempty"`
	Flows    map[string][]Config `toml:"flows,omitempty"`
	Outlets  map[string][]Config `toml:"outlets,omitempty"`
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

func (c Config) GetConfigArray(key string, defaultVal []Config) []Config {
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

func (c Config) GetStringArray(key string, defaultVal []string) []string {
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

func (c Config) GetIntArray(key string, defaultVal []int) []int {
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
