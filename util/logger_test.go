package util

import "testing"

func ExampleNewLogger() {
	// This example demonstrates how to create a new logger with different log levels.
	// The logger is configured to write to a file and the log level is set to INFO.
	conf := LogConfig{
		Filename: "-", // "-" writes to stdout, "" discard logging, otherwise writes to a file
		Level:    LOG_LEVEL_INFO,
	}
	logger := NewLogger(conf)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}

func TestNewLogger(t *testing.T) {
	// This test demonstrates how to test the NewLogger function.
	conf := LogConfig{
		Filename: "-",
		Level:    LOG_LEVEL_DEBUG,
	}
	logger := NewLogger(conf)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}
