// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
	}{
		{"simple version", "1.2.3", [3]int{1, 2, 3}},
		{"with v prefix", "v1.2.3", [3]int{1, 2, 3}},
		{"major only", "1", [3]int{1, 0, 0}},
		{"major.minor", "1.2", [3]int{1, 2, 0}},
		{"large numbers", "10.20.30", [3]int{10, 20, 30}},
		{"zero version", "0.0.0", [3]int{0, 0, 0}},
		{"empty version", "", [3]int{0, 0, 0}},
		{"invalid chars", "abc", [3]int{0, 0, 0}},
		{"partial invalid", "1.2.abc", [3]int{1, 2, 0}},
		{"extra parts", "1.2.3.4", [3]int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestIsVersionCompatible(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		minVersion string
		expected   bool
	}{
		{"equal versions", "1.0.0", "1.0.0", true},
		{"higher major", "2.0.0", "1.0.0", true},
		{"higher minor", "1.1.0", "1.0.0", true},
		{"higher patch", "1.0.1", "1.0.0", true},
		{"lower major", "0.9.0", "1.0.0", false},
		{"lower minor", "1.0.0", "1.1.0", false},
		{"lower patch", "1.0.0", "1.0.1", false},
		{"with v prefix", "v3.0.1", "1.0.0", true},
		{"real version check", "3.0.1", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVersionCompatible(tt.version, tt.minVersion)
			if result != tt.expected {
				t.Errorf("isVersionCompatible(%q, %q) = %v, want %v",
					tt.version, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestMinimumWarpgateVersion(t *testing.T) {
	// Verify the constant is set correctly
	if MinimumWarpgateVersion != "3.0.0" {
		t.Errorf("MinimumWarpgateVersion = %q, want %q", MinimumWarpgateVersion, "3.0.0")
	}
}

func TestWarpgateClientMethods(t *testing.T) {
	// Test that client methods don't panic with nil/empty values
	client := &WarpgateClient{}

	// These should return empty strings, not panic
	if client.GetRepoPath() != "" {
		t.Errorf("GetRepoPath() on empty client should return empty string")
	}
	if client.GetBinaryPath() != "" {
		t.Errorf("GetBinaryPath() on empty client should return empty string")
	}
	if client.GetCLIVersion() != "" {
		t.Errorf("GetCLIVersion() on empty client should return empty string")
	}
	if client.IsCLIAvailable() != false {
		t.Errorf("IsCLIAvailable() on empty client should return false")
	}
}

func TestBuildOptions(t *testing.T) {
	// Test that BuildOptions struct can be created with all fields
	opts := BuildOptions{
		Template:      "test-template",
		Target:        "container",
		Architectures: []string{"amd64", "arm64"},
		Push:          true,
		Registry:      "ghcr.io/test",
		Vars:          map[string]string{"key": "value"},
		Tags:          []string{"latest", "v1.0.0"},
		NoCache:       true,
		SaveDigests:   true,
		DigestDir:     "/tmp/digests",
	}

	if opts.Template != "test-template" {
		t.Errorf("BuildOptions.Template = %q, want %q", opts.Template, "test-template")
	}
	if len(opts.Architectures) != 2 {
		t.Errorf("BuildOptions.Architectures length = %d, want 2", len(opts.Architectures))
	}
}

func TestBuildArgs(t *testing.T) {
	// buildArgs converts BuildOptions to the argv warpgate expects. With 30+
	// optional fields it's easy to forget a passthrough — assert each block's
	// representative flags so silent omissions get caught.
	opts := BuildOptions{
		Template:      "attack-box",
		FromGit:       "https://example.com/repo.git",
		Target:        "ami",
		Architectures: []string{"amd64", "arm64"},
		Push:          true,
		PushDigest:    true,
		Registry:      "ghcr.io/test",
		Vars:          map[string]string{"key": "value"},
		VarFiles:      []string{"vars.yaml"},
		BuildArgs:     []string{"FOO=bar"},
		Tags:          []string{"latest"},
		NoCache:       true,
		SaveDigests:   true,
		DigestDir:     "/tmp/digests",

		Region:          "us-west-2",
		InstanceType:    "t3.large",
		Force:           true,
		Cleanup:         true,
		DryRun:          true,
		Regions:         []string{"us-west-2", "us-east-1"},
		ParallelRegions: true,
		CopyToRegions:   []string{"eu-west-1"},
		StreamLogs:      true,
		ShowEC2Status:   true,
		OutputManifest:  "manifest.json",

		AzureSubscription: "sub-123",
		AzureLocation:     "eastus",
		AzureResourceGrp:  "rg-build",
		AzureGallery:      "gallery",
		AzureImageDef:     "imgdef",
		AzureVMSize:       "Standard_D2s_v5",
		AzureIdentityID:   "/subscriptions/.../identities/build",
		AzureTargetRegion: []string{"westus", "northeurope"},
		AzureSubnetID:     "/subnet/id",
		AzureProxyVMSize:  "Standard_B1s",

		ProxmoxEndpoint: "https://proxmox.example:8006",
		ProxmoxNode:     "pve01",
		ProxmoxStorage:  "local-lvm",
		ProxmoxPool:     "build-pool",
	}

	args := buildArgs("", opts)
	joined := strings.Join(args, " ")

	wantFlags := []string{
		"build",
		"--template attack-box",
		"--from-git https://example.com/repo.git",
		"--target ami",
		"--arch amd64", "--arch arm64",
		"--push",
		"--push-digest",
		"--registry ghcr.io/test",
		"--var key=value",
		"--var-file vars.yaml",
		"--build-arg FOO=bar",
		"--tag latest",
		"--no-cache",
		"--save-digests",
		"--digest-dir /tmp/digests",
		"--region us-west-2",
		"--instance-type t3.large",
		"--force",
		"--cleanup",
		"--dry-run",
		"--regions us-west-2", "--regions us-east-1",
		"--parallel-regions",
		"--copy-to-regions eu-west-1",
		"--stream-logs",
		"--show-ec2-status",
		"--output-manifest manifest.json",
		"--subscription sub-123",
		"--location eastus",
		"--resource-group rg-build",
		"--gallery gallery",
		"--image-definition imgdef",
		"--vm-size Standard_D2s_v5",
		"--identity-id /subscriptions/.../identities/build",
		"--target-regions westus", "--target-regions northeurope",
		"--subnet-id /subnet/id",
		"--proxy-vm-size Standard_B1s",
		"--proxmox-endpoint https://proxmox.example:8006",
		"--proxmox-node pve01",
		"--proxmox-storage local-lvm",
		"--proxmox-pool build-pool",
	}
	for _, f := range wantFlags {
		if !strings.Contains(joined, f) {
			t.Errorf("buildArgs missing %q in: %s", f, joined)
		}
	}
}

func TestInitOptions(t *testing.T) {
	opts := InitOptions{
		OutputDir:    "/tmp/output",
		FromTemplate: "base-template",
	}

	if opts.OutputDir != "/tmp/output" {
		t.Errorf("InitOptions.OutputDir = %q, want %q", opts.OutputDir, "/tmp/output")
	}
	if opts.FromTemplate != "base-template" {
		t.Errorf("InitOptions.FromTemplate = %q, want %q", opts.FromTemplate, "base-template")
	}
}

func TestNewWarpgateClientWithValidPath(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	client, err := NewWarpgateClient(tmpDir)
	if err != nil {
		t.Fatalf("NewWarpgateClient failed: %v", err)
	}

	if client.GetRepoPath() != tmpDir {
		t.Errorf("GetRepoPath() = %q, want %q", client.GetRepoPath(), tmpDir)
	}
}

func TestNewWarpgateClientWithEmptyPath(t *testing.T) {
	// Test with empty path - should succeed with no repoPath set
	client, err := NewWarpgateClient("")
	if err != nil {
		t.Fatalf("NewWarpgateClient with empty path failed: %v", err)
	}

	if client.GetRepoPath() != "" {
		t.Errorf("GetRepoPath() = %q, want empty string", client.GetRepoPath())
	}
}

func TestNewWarpgateClientWithNonExistentPath(t *testing.T) {
	_, err := NewWarpgateClient("/nonexistent/path/12345")
	if err == nil {
		t.Error("NewWarpgateClient should fail for non-existent path")
	}
}

func TestExecuteCLINotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.ExecuteCLI(context.Background(), "version")
	if err == nil {
		t.Error("ExecuteCLI should fail when CLI is not available")
	}
}

func TestExecuteCLIWithWorkdirNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.ExecuteCLIWithWorkdir(context.Background(), "/tmp", "version")
	if err == nil {
		t.Error("ExecuteCLIWithWorkdir should fail when CLI is not available")
	}
}

func TestWarpgateBuildArgsConstruction(t *testing.T) {
	// Test that WarpgateBuild constructs arguments correctly
	// by checking it fails appropriately when CLI is not available
	client := &WarpgateClient{
		cliDetected: false,
	}

	opts := BuildOptions{
		Target:        "container",
		Architectures: []string{"amd64", "arm64"},
		Push:          true,
		Registry:      "ghcr.io/test",
		Vars:          map[string]string{"key": "value"},
		Tags:          []string{"latest"},
		NoCache:       true,
		SaveDigests:   true,
		DigestDir:     "/tmp/digests",
	}

	_, err := client.WarpgateBuild(context.Background(), "test-template", opts)
	if err == nil {
		t.Error("WarpgateBuild should fail when CLI is not available")
	}
}

func TestWarpgateValidateNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateValidate(context.Background(), "/path/to/config", true)
	if err == nil {
		t.Error("WarpgateValidate should fail when CLI is not available")
	}
}

func TestWarpgateInitNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	opts := InitOptions{
		OutputDir:    "/tmp/output",
		FromTemplate: "base",
	}

	_, err := client.WarpgateInit(context.Background(), "test", opts)
	if err == nil {
		t.Error("WarpgateInit should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesListNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateTemplatesList(context.Background(), "local", "json", false)
	if err == nil {
		t.Error("WarpgateTemplatesList should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesInfoNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateTemplatesInfo(context.Background(), "test-template")
	if err == nil {
		t.Error("WarpgateTemplatesInfo should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesAddNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateTemplatesAdd(context.Background(), "https://example.com/template", "my-template")
	if err == nil {
		t.Error("WarpgateTemplatesAdd should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesRemoveNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateTemplatesRemove(context.Background(), "my-template")
	if err == nil {
		t.Error("WarpgateTemplatesRemove should fail when CLI is not available")
	}
}

func TestWarpgateManifestsCreateNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateManifestsCreate(context.Background(), ManifestsCreateOptions{
		Name:     "test-manifest",
		Registry: "ghcr.io/test",
	})
	if err == nil {
		t.Error("WarpgateManifestsCreate should fail when CLI is not available")
	}
}

func TestWarpgateConfigGetNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateConfigGet(context.Background(), "registry")
	if err == nil {
		t.Error("WarpgateConfigGet should fail when CLI is not available")
	}
}

func TestWarpgateConfigSetNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateConfigSet(context.Background(), "registry", "ghcr.io/test")
	if err == nil {
		t.Error("WarpgateConfigSet should fail when CLI is not available")
	}
}

func TestWarpgateConfigShowNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateConfigShow(context.Background())
	if err == nil {
		t.Error("WarpgateConfigShow should fail when CLI is not available")
	}
}

func TestWarpgateConvertNotAvailable(t *testing.T) {
	client := &WarpgateClient{
		cliDetected: false,
	}

	_, err := client.WarpgateConvert(context.Background(), ConvertOptions{
		Source: "/path/to/source",
		Output: "/path/to/output",
	})
	if err == nil {
		t.Error("WarpgateConvert should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesSearchNotAvailable(t *testing.T) {
	c := &WarpgateClient{cliDetected: false}
	if _, err := c.WarpgateTemplatesSearch(context.Background(), "attack"); err == nil {
		t.Error("WarpgateTemplatesSearch should fail when CLI is not available")
	}
}

func TestWarpgateTemplatesUpdateNotAvailable(t *testing.T) {
	c := &WarpgateClient{cliDetected: false}
	if _, err := c.WarpgateTemplatesUpdate(context.Background()); err == nil {
		t.Error("WarpgateTemplatesUpdate should fail when CLI is not available")
	}
}

func TestWarpgateConfigInitNotAvailable(t *testing.T) {
	c := &WarpgateClient{cliDetected: false}
	if _, err := c.WarpgateConfigInit(context.Background(), false); err == nil {
		t.Error("WarpgateConfigInit should fail when CLI is not available")
	}
}

func TestWarpgateConfigPathNotAvailable(t *testing.T) {
	c := &WarpgateClient{cliDetected: false}
	if _, err := c.WarpgateConfigPath(context.Background()); err == nil {
		t.Error("WarpgateConfigPath should fail when CLI is not available")
	}
}

func TestWarpgateCleanupNotAvailable(t *testing.T) {
	c := &WarpgateClient{cliDetected: false}
	if _, err := c.WarpgateCleanup(context.Background(), CleanupOptions{DryRun: true}); err == nil {
		t.Error("WarpgateCleanup should fail when CLI is not available")
	}
}

func TestNewWarpgateClientWithBinaryInvalidPath(t *testing.T) {
	_, err := NewWarpgateClientWithBinary("", "/nonexistent/warpgate/binary")
	if err == nil {
		t.Error("NewWarpgateClientWithBinary should fail with non-existent binary path")
	}
}

func TestNewWarpgateClientWithBinaryInvalidRepoPath(t *testing.T) {
	// Create a fake binary file
	tmpFile, err := os.CreateTemp("", "warpgate-fake-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// This will fail because the binary can't be executed for version check
	_, err = NewWarpgateClientWithBinary("/nonexistent/repo", tmpFile.Name())
	if err == nil {
		t.Error("NewWarpgateClientWithBinary should fail with invalid binary")
	}
}
