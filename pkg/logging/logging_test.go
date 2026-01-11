// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package logging

import (
	"os"
	"strings"
	"testing"
)

func TestNewLoggerWithFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "warpgate-log-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	logger, err := NewLogger(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Test logging
	logger.Infof("test info message")
	logger.Errorf("test error message")
	logger.Debugf("test debug message")

	// Read log file and verify content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test info message") {
		t.Error("Log file should contain info message")
	}
	if !strings.Contains(contentStr, "test error message") {
		t.Error("Log file should contain error message")
	}
}

func TestNewLoggerWithEmptyPath(t *testing.T) {
	// Empty path should use stderr (not create file)
	logger, err := NewLogger("")
	if err != nil {
		t.Fatalf("Failed to create logger with empty path: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Should not panic when logging
	logger.Infof("test message to stderr")
	logger.Errorf("test error to stderr")
}

func TestLoggerWriter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "warpgate-log-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	logger, err := NewLogger(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	writer := logger.Writer()
	if writer == nil {
		t.Fatal("Logger.Writer() should not return nil")
	}

	// Write to the writer
	_, err = writer.Write([]byte("direct write test\n"))
	if err != nil {
		t.Errorf("Failed to write to logger writer: %v", err)
	}
}

func TestLoggerInvalidPath(t *testing.T) {
	// Try to create logger with invalid path
	_, err := NewLogger("/nonexistent/directory/that/does/not/exist/log.txt")
	if err == nil {
		t.Error("Expected error when creating logger with invalid path")
	}
}
