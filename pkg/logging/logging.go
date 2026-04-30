// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

// Package logging provides structured logging utilities for the warpgate-mcp-server.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
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
		file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) //nolint:gosec // G302: log files need to be readable by operators
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

// GetLogDir returns the appropriate log directory based on the operating system
// - macOS: ~/Library/Logs/warpgate-mcp-server/
// - Linux: ~/.local/share/warpgate-mcp-server/logs/ (XDG_DATA_HOME)
// - Windows: %LOCALAPPDATA%/warpgate-mcp-server/logs/
func GetLogDir() (string, error) {
	var logDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Logs/warpgate-mcp-server/
		logDir = filepath.Join(homeDir, "Library", "Logs", "warpgate-mcp-server")
	case "linux":
		// Linux: Follow XDG Base Directory specification
		// Use XDG_DATA_HOME if set, otherwise ~/.local/share
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			xdgDataHome = filepath.Join(homeDir, ".local", "share")
		}
		logDir = filepath.Join(xdgDataHome, "warpgate-mcp-server", "logs")
	case "windows":
		// Windows: %LOCALAPPDATA%/warpgate-mcp-server/logs/
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		logDir = filepath.Join(localAppData, "warpgate-mcp-server", "logs")
	default:
		// Fallback for unknown OS
		logDir = filepath.Join(homeDir, ".warpgate-mcp-server", "logs")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0750); err != nil { //nolint:gosec // G703: logDir is derived from controlled env vars and a fixed subpath
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	return logDir, nil
}

// CreateBuildLogFile creates a new log file for a build with a timestamp
// Returns the file path and an open file handle
func CreateBuildLogFile(template string) (*os.File, string, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return nil, "", err
	}

	// Create builds subdirectory
	buildsDir := filepath.Join(logDir, "builds")
	if err := os.MkdirAll(buildsDir, 0750); err != nil {
		return nil, "", fmt.Errorf("failed to create builds log directory: %w", err)
	}

	// Create log filename with timestamp and template name
	// Clean template name to be filesystem-safe
	safeName := filepath.Base(template)
	if safeName == "." || safeName == "/" {
		safeName = "unnamed"
	}
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s_%s.log", safeName, timestamp)
	logPath := filepath.Join(buildsDir, logFileName)

	// Open log file for writing
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600) //nolint:gosec // G304: logPath is constructed from controlled directory and sanitized template name
	if err != nil {
		return nil, "", fmt.Errorf("failed to create build log file: %w", err)
	}

	return file, logPath, nil
}
