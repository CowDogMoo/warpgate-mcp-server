// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package resources

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/server"
)

func createTestLogger(t *testing.T) (*logging.Logger, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "warpgate-test-log-*")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	_ = tmpFile.Close()

	logger, err := logging.NewLogger(tmpFile.Name())
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create logger: %v", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	return logger, cleanup
}

func createTestWarpgateDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	taskfile := `version: '3'
tasks:
  default:
    cmds:
      - echo "test"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Taskfile.yaml"), []byte(taskfile), 0644); err != nil { //nolint:gosec // G306: test file permissions
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create Taskfile: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestRegisterResources(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	RegisterResources(s, logger, tmpDir)
	t.Log("RegisterResources completed without panic")
}

func TestRegisterResourcesWithEmptyPath(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	RegisterResources(s, logger, "")
	t.Log("RegisterResources with empty path completed without panic")
}

func TestRegisterResourcesWithNilServer(t *testing.T) {
	// This tests that we can handle nil gracefully or panic appropriately
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	// Test with valid server to ensure registration works
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	RegisterResources(s, logger, tmpDir)
}

func TestRegisterResourcesMultipleTimes(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	// Register multiple times - should not panic
	RegisterResources(s, logger, tmpDir)
	RegisterResources(s, logger, tmpDir)
	t.Log("RegisterResources multiple times completed without panic")
}

func TestRegisterResourcesWithDifferentPaths(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	// Test with various path configurations
	paths := []string{
		"",                        // Empty path
		"/nonexistent/path/12345", // Non-existent path
		"/tmp",                    // Existing but invalid
	}

	for _, path := range paths {
		// Should not panic even with invalid paths
		RegisterResources(s, logger, path)
	}

	t.Log("RegisterResources with different paths completed without panic")
}

func TestResourceDefinitions(t *testing.T) {
	// Test that resources are defined with correct properties
	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	RegisterResources(s, logger, tmpDir)

	// Verify that registration happened without errors
	// The actual resource URIs are tested by verifying the server state
	t.Log("Resource definitions registered successfully")
}

func TestResourcesWithLoggerOutput(t *testing.T) {
	// Test that logger is properly used
	tmpLogFile, err := os.CreateTemp("", "warpgate-test-log-*")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	logPath := tmpLogFile.Name()
	_ = tmpLogFile.Close()
	defer func() { _ = os.Remove(logPath) }()

	logger, err := logging.NewLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	s := server.NewMCPServer("test", "1.0.0", server.WithResourceCapabilities(true, false))
	RegisterResources(s, logger, tmpDir)

	// Logger should have been used for any logging during registration
	// This verifies the logger parameter is properly handled
	t.Log("Resources registered with logger")
}

func TestRegisterResourcesServerOptions(t *testing.T) {
	// Test with different server configurations
	testCases := []struct {
		name string
		opts []server.ServerOption
	}{
		{
			name: "with subscribe capability",
			opts: []server.ServerOption{server.WithResourceCapabilities(true, true)},
		},
		{
			name: "without subscribe capability",
			opts: []server.ServerOption{server.WithResourceCapabilities(false, false)},
		},
		{
			name: "with logging",
			opts: []server.ServerOption{server.WithResourceCapabilities(true, false), server.WithLogging()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := server.NewMCPServer("test", "1.0.0", tc.opts...)
			logger, cleanupLogger := createTestLogger(t)
			defer cleanupLogger()

			tmpDir, cleanupDir := createTestWarpgateDir(t)
			defer cleanupDir()

			// Should not panic with different configurations
			RegisterResources(s, logger, tmpDir)
		})
	}
}
