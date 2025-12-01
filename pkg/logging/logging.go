// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with MCP-specific functionality
type Logger struct {
	*slog.Logger
	writer io.Writer
}

// NewLogger creates a new Logger instance
// If outPath is empty, logs to stderr
// Otherwise, logs to the specified file
func NewLogger(outPath string) (*Logger, error) {
	var writer io.Writer
	var handler slog.Handler

	if outPath == "" {
		writer = os.Stderr
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := &Logger{
		Logger: slog.New(handler),
		writer: writer,
	}

	return logger, nil
}

// Writer returns the underlying writer
func (l *Logger) Writer() io.Writer {
	return l.writer
}

// Infof logs an informational message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}
