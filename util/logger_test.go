package util

import "testing"

func ExampleNewLogger() {
	// This example demonstrates how to create a new logger with different log levels.
	// The logger is configured to write to a file and the log level is set to INFO.
	conf := LogConfig{
		Path:       "-", // "-" writes to stdout, "" discard logging, otherwise writes to a file
		Level:      LOG_LEVEL_WARN,
		NoColor:    true,
		Timeformat: "no-time-for-test",
	}
	logger := NewLogger(conf)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
	// Output:
	// no-time-for-test WRN This is a warning message
	// no-time-for-test ERR This is an error message
}

func TestNewLogger(t *testing.T) {
	// This test demonstrates how to test the NewLogger function.
	conf := LogConfig{
		Path:  "-",
		Level: LOG_LEVEL_DEBUG,
	}
	logger := NewLogger(conf)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}
