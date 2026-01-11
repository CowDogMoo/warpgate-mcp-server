// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

//go:build integration

package client

import (
	"os"
	"testing"
)

// Integration tests that require warpgate CLI to be installed
// Run with: go test -tags=integration ./...

func TestNewWarpgateClientWithCLI(t *testing.T) {
	// Skip if warpgate is not installed
	client := &WarpgateClient{}
	if err := client.detectWarpgateBinary(""); err != nil {
		t.Skip("Skipping: warpgate CLI not installed")
	}

	// Test that we can detect the CLI
	if !client.IsCLIAvailable() {
		t.Error("CLI should be available after successful detection")
	}

	version := client.GetCLIVersion()
	if version == "" {
		t.Error("CLI version should not be empty")
	}
	t.Logf("Detected warpgate version: %s", version)

	binaryPath := client.GetBinaryPath()
	if binaryPath == "" {
		t.Error("Binary path should not be empty")
	}
	t.Logf("Detected binary path: %s", binaryPath)
}

func TestWarpgateTemplatesListIntegration(t *testing.T) {
	client := &WarpgateClient{}
	if err := client.detectWarpgateBinary(""); err != nil {
		t.Skip("Skipping: warpgate CLI not installed")
	}

	output, err := client.WarpgateTemplatesList("", "json")
	if err != nil {
		t.Logf("Templates list returned error (may be expected): %v", err)
		return
	}
	t.Logf("Templates list output: %s", output)
}

func TestWarpgateConfigShowIntegration(t *testing.T) {
	client := &WarpgateClient{}
	if err := client.detectWarpgateBinary(""); err != nil {
		t.Skip("Skipping: warpgate CLI not installed")
	}

	output, err := client.WarpgateConfigShow()
	if err != nil {
		t.Logf("Config show returned error (may be expected): %v", err)
		return
	}
	t.Logf("Config output: %s", output)
}

func TestNewWarpgateClientWithRepoPath(t *testing.T) {
	// Create temp directory with Taskfile
	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := `version: '3'
tasks:
  default:
    cmds:
      - echo "test"
`
	if err := os.WriteFile(tmpDir+"/Taskfile.yaml", []byte(taskfile), 0644); err != nil {
		t.Fatalf("Failed to create Taskfile: %v", err)
	}

	client, err := NewWarpgateClient(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.GetRepoPath() != tmpDir {
		t.Errorf("GetRepoPath() = %q, want %q", client.GetRepoPath(), tmpDir)
	}
}
