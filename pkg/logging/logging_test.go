// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package logging

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestGetLogDir(t *testing.T) {
	logDir, err := GetLogDir()
	if err != nil {
		t.Fatalf("GetLogDir() failed: %v", err)
	}

	if logDir == "" {
		t.Fatal("GetLogDir() returned empty string")
	}

	// Verify directory was created
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("Log directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("Log directory path is not a directory")
	}

	// Verify it follows OS conventions
	homeDir, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		expectedPath := filepath.Join(homeDir, "Library", "Logs", "warpgate-mcp-server")
		if logDir != expectedPath {
			t.Errorf("Expected log dir %s, got %s", expectedPath, logDir)
		}
	case "linux":
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			xdgDataHome = filepath.Join(homeDir, ".local", "share")
		}
		expectedPath := filepath.Join(xdgDataHome, "warpgate-mcp-server", "logs")
		if logDir != expectedPath {
			t.Errorf("Expected log dir %s, got %s", expectedPath, logDir)
		}
	}
}

func TestCreateBuildLogFile(t *testing.T) {
	template := "test-template"
	file, logPath, err := CreateBuildLogFile(template)
	if err != nil {
		t.Fatalf("CreateBuildLogFile() failed: %v", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(logPath)
	}()

	// Verify file was created
	if logPath == "" {
		t.Fatal("CreateBuildLogFile() returned empty log path")
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Log file was not created: %v", err)
	}

	if info.IsDir() {
		t.Fatal("Log path is a directory, not a file")
	}

	// Verify file contains template name
	if !strings.Contains(filepath.Base(logPath), template) {
		t.Errorf("Log filename should contain template name '%s', got '%s'", template, filepath.Base(logPath))
	}

	// Verify we can write to the file
	testLine := "Test build output\n"
	_, err = file.WriteString(testLine)
	if err != nil {
		t.Fatalf("Failed to write to log file: %v", err)
	}

	// Close and read back
	_ = file.Close()
	content, err := os.ReadFile(logPath) //nolint:gosec // G304: Reading test file we just created
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if string(content) != testLine {
		t.Errorf("Expected log content '%s', got '%s'", testLine, string(content))
	}
}

func TestCreateBuildLogFileWithComplexTemplate(t *testing.T) {
	// Test with path-like template name
	template := "/path/to/my-template"
	file, logPath, err := CreateBuildLogFile(template)
	if err != nil {
		t.Fatalf("CreateBuildLogFile() failed: %v", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(logPath)
	}()

	// Should only use the basename
	basename := filepath.Base(logPath)
	if !strings.Contains(basename, "my-template") {
		t.Errorf("Log filename should contain 'my-template', got '%s'", basename)
	}

	// Should not contain full path
	if strings.Contains(basename, "/path/to/") {
		t.Errorf("Log filename should not contain full path, got '%s'", basename)
	}
}
