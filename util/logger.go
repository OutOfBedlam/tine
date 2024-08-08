package util

import (
	"io"
	"log/slog"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LOG_LEVEL_DEBUG = "DEBUG"
	LOG_LEVEL_INFO  = "INFO"
	LOG_LEVEL_WARN  = "WARN"
	LOG_LEVEL_ERROR = "ERROR"
)

type LogConfig struct {
	Path       string `toml:"path"`
	NoColor    bool   `toml:"no_color,omitempty"`
	AddSource  bool   `toml:"add_source,omitempty"`
	Timeformat string `toml:"timeformat,omitempty"`
	Level      string `toml:"level"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
	Compress   bool   `toml:"compress,omitempty"`
	Chown      string `toml:"chown,omitempty"`
}

func NewLogger(conf LogConfig) *slog.Logger {
	var lw io.Writer
	if conf.Timeformat == "" {
		conf.Timeformat = "2006-01-02 15:04:05 -0700"
	}
	opt := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: conf.Timeformat,
		AddSource:  conf.AddSource,
		NoColor:    conf.NoColor,
	}
	switch strings.ToUpper(conf.Level) {
	case LOG_LEVEL_DEBUG:
		opt.Level = slog.LevelDebug
	case LOG_LEVEL_INFO:
		opt.Level = slog.LevelInfo
	case LOG_LEVEL_WARN:
		opt.Level = slog.LevelWarn
	case LOG_LEVEL_ERROR:
		opt.Level = slog.LevelError
	}

	if conf.Path == "" {
		// disable logging
		lw = io.Discard
	} else if conf.Path == "-" {
		lw = os.Stdout
	} else {
		lw = &lumberjack.Logger{
			Filename:   conf.Path,
			MaxSize:    conf.MaxSize,
			MaxAge:     conf.MaxAge,
			MaxBackups: conf.MaxBackups,
			LocalTime:  true,
			Compress:   conf.Compress,
		}
	}
	ll := tint.NewHandler(lw, opt)
	ret := slog.New(ll)
	if conf.Chown == "" || runtime.GOOS == "windows" {
		return ret
	}
	usr, err := user.Lookup(conf.Chown)
	if err != nil {
		slog.Error("failed to lookup user for chown log file", "error", err.Error())
		return ret
	}
	uid, err := strconv.ParseInt(usr.Uid, 10, 32)
	if err != nil {
		slog.Error("failed to parse uid for chown log file", "error", err.Error())
		return ret
	}
	gid, err := strconv.ParseInt(usr.Gid, 10, 32)
	if err != nil {
		slog.Error("failed to parse gid for chown log file", "error", err.Error())
		return ret
	}
	if err := os.Chown(conf.Path, int(uid), int(gid)); err != nil {
		slog.Error("failed to chown log file", "error", err.Error())
	}
	return ret
}
