// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

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

func TestRegisterTools(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	RegisterTools(s, logger, tmpDir)
	t.Log("RegisterTools completed without panic")
}

func TestRegisterToolsWithEmptyPath(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	RegisterTools(s, logger, "")
	t.Log("RegisterTools with empty path completed without panic")
}

func TestRegisterToolsWithInvalidPath(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	RegisterTools(s, logger, "/nonexistent/path/12345")
	t.Log("RegisterTools with invalid path completed without panic")
}

func TestRegisterToolsMultipleTimes(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	// Should not panic when called multiple times
	RegisterTools(s, logger, tmpDir)
	RegisterTools(s, logger, tmpDir)
	t.Log("RegisterTools multiple times completed without panic")
}

func TestRegisterToolsWithDifferentServerOptions(t *testing.T) {
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	testCases := []struct {
		name string
		opts []server.ServerOption
	}{
		{
			name: "basic server",
			opts: []server.ServerOption{},
		},
		{
			name: "with logging",
			opts: []server.ServerOption{server.WithLogging()},
		},
		{
			name: "with resources",
			opts: []server.ServerOption{server.WithResourceCapabilities(true, false)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			s := server.NewMCPServer("test", "1.0.0", tc.opts...)
			RegisterTools(s, logger, tmpDir)
		})
	}
}

func TestIndividualToolRegistration(t *testing.T) {
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	tmpDir, cleanupDir := createTestWarpgateDir(t)
	defer cleanupDir()

	// Test each tool registration function individually
	t.Run("warpgateBuild", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateBuild(s, logger, tmpDir)
	})

	t.Run("warpgateValidate", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateValidate(s, logger, tmpDir)
	})

	t.Run("warpgateInit", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateInit(s, logger, tmpDir)
	})

	t.Run("warpgateTemplatesList", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateTemplatesList(s, logger, tmpDir)
	})

	t.Run("warpgateTemplatesInfo", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateTemplatesInfo(s, logger, tmpDir)
	})

	t.Run("warpgateTemplatesAdd", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateTemplatesAdd(s, logger, tmpDir)
	})

	t.Run("warpgateTemplatesRemove", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateTemplatesRemove(s, logger, tmpDir)
	})

	t.Run("createTemplate", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		createTemplate(s, logger, tmpDir)
	})

	t.Run("warpgateManifestsCreate", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateManifestsCreate(s, logger, tmpDir)
	})

	t.Run("warpgateManifestsPush", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateManifestsPush(s, logger, tmpDir)
	})

	t.Run("warpgateConfigGet", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateConfigGet(s, logger, tmpDir)
	})

	t.Run("warpgateConfigSet", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateConfigSet(s, logger, tmpDir)
	})

	t.Run("warpgateConfigShow", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateConfigShow(s, logger, tmpDir)
	})

	t.Run("warpgateConvert", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		warpgateConvert(s, logger, tmpDir)
	})

	t.Run("listTasks", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		listTasks(s, logger, tmpDir)
	})

	t.Run("runTask", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		runTask(s, logger, tmpDir)
	})

	t.Run("runImageBuilder", func(_ *testing.T) {
		s := server.NewMCPServer("test", "1.0.0")
		runImageBuilder(s, logger, tmpDir)
	})
}

func TestToolRegistrationWithEmptyPath(t *testing.T) {
	logger, cleanupLogger := createTestLogger(t)
	defer cleanupLogger()

	// All tools should register without panicking even with empty path
	s := server.NewMCPServer("test", "1.0.0")

	warpgateBuild(s, logger, "")
	warpgateValidate(s, logger, "")
	warpgateInit(s, logger, "")
	warpgateTemplatesList(s, logger, "")
	warpgateTemplatesInfo(s, logger, "")
	warpgateTemplatesAdd(s, logger, "")
	warpgateTemplatesRemove(s, logger, "")
	createTemplate(s, logger, "")
	warpgateManifestsCreate(s, logger, "")
	warpgateManifestsPush(s, logger, "")
	warpgateConfigGet(s, logger, "")
	warpgateConfigSet(s, logger, "")
	warpgateConfigShow(s, logger, "")
	warpgateConvert(s, logger, "")
	listTasks(s, logger, "")
	runTask(s, logger, "")
	runImageBuilder(s, logger, "")

	t.Log("All tools registered with empty path without panic")
}
