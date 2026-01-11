// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// MinimumWarpgateVersion is the minimum supported warpgate CLI version
const MinimumWarpgateVersion = "1.0.0"

// WarpgateClient handles interaction with the Warpgate repository and CLI
type WarpgateClient struct {
	repoPath    string
	binaryPath  string
	cliVersion  string
	cliDetected bool
}

// NewWarpgateClient creates a new Warpgate client
// If repoPath is empty, it will attempt to find the warpgate repo
func NewWarpgateClient(repoPath string) (*WarpgateClient, error) {
	client := &WarpgateClient{}

	// Detect warpgate CLI binary
	if err := client.detectWarpgateBinary(""); err != nil {
		// CLI detection failed, but we can still use task-based operations
		client.cliDetected = false
	} else {
		client.cliDetected = true
	}

	if repoPath == "" {
		// Try to auto-detect the warpgate repository
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		// Common locations for the warpgate repo
		possiblePaths := []string{
			filepath.Join(home, "warpgate"),
			filepath.Join(home, "cowdogmoo", "warpgate"),
			filepath.Join(home, "code", "warpgate"),
			filepath.Join(home, "projects", "warpgate"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(filepath.Join(path, "Taskfile.yaml")); err == nil {
				repoPath = path
				break
			}
		}

		if repoPath == "" {
			return nil, fmt.Errorf("warpgate repository not found in common locations")
		}
	}

	// Verify the path exists and contains a Taskfile
	taskfilePath := filepath.Join(repoPath, "Taskfile.yaml")
	if _, err := os.Stat(taskfilePath); err != nil {
		return nil, fmt.Errorf("Taskfile.yaml not found in %s: %w", repoPath, err)
	}

	client.repoPath = repoPath
	return client, nil
}

// NewWarpgateClientWithBinary creates a client with explicit binary path
func NewWarpgateClientWithBinary(repoPath, binaryPath string) (*WarpgateClient, error) {
	client := &WarpgateClient{}

	// Detect warpgate CLI with explicit path
	if err := client.detectWarpgateBinary(binaryPath); err != nil {
		return nil, fmt.Errorf("warpgate binary detection failed: %w", err)
	}
	client.cliDetected = true

	if repoPath != "" {
		// Verify the path exists and contains a Taskfile
		taskfilePath := filepath.Join(repoPath, "Taskfile.yaml")
		if _, err := os.Stat(taskfilePath); err != nil {
			return nil, fmt.Errorf("Taskfile.yaml not found in %s: %w", repoPath, err)
		}
		client.repoPath = repoPath
	}

	return client, nil
}

// detectWarpgateBinary finds the warpgate binary and validates its version
func (w *WarpgateClient) detectWarpgateBinary(explicitPath string) error {
	var binaryPath string

	if explicitPath != "" {
		// Use explicit path
		if _, err := os.Stat(explicitPath); err != nil {
			return fmt.Errorf("warpgate binary not found at %s: %w", explicitPath, err)
		}
		binaryPath = explicitPath
	} else {
		// Search in PATH first
		if path, err := exec.LookPath("warpgate"); err == nil {
			binaryPath = path
		} else {
			// Try common install locations
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			commonPaths := []string{
				filepath.Join(home, "go", "bin", "warpgate"),
				"/usr/local/bin/warpgate",
				"/usr/bin/warpgate",
				filepath.Join(home, ".local", "bin", "warpgate"),
			}

			for _, path := range commonPaths {
				if _, err := os.Stat(path); err == nil {
					binaryPath = path
					break
				}
			}
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("warpgate binary not found in PATH or common locations")
	}

	// Validate version
	version, err := w.getWarpgateVersion(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to get warpgate version: %w", err)
	}

	if !isVersionCompatible(version, MinimumWarpgateVersion) {
		return fmt.Errorf("warpgate version %s is below minimum required version %s",
			version, MinimumWarpgateVersion)
	}

	w.binaryPath = binaryPath
	w.cliVersion = version
	return nil
}

// getWarpgateVersion extracts the version from warpgate CLI
func (w *WarpgateClient) getWarpgateVersion(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "version", "--json")
	output, err := cmd.Output()
	if err != nil {
		// Try without --json flag
		cmd = exec.Command(binaryPath, "--version")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to execute warpgate version: %w", err)
		}
		// Parse version from output like "warpgate version v3.0.1"
		re := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
		matches := re.FindStringSubmatch(string(output))
		if len(matches) < 2 {
			return "", fmt.Errorf("could not parse version from output: %s", string(output))
		}
		return matches[1], nil
	}

	// Parse JSON output
	var versionInfo struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		return "", fmt.Errorf("failed to parse version JSON: %w", err)
	}
	// Strip 'v' prefix if present
	return strings.TrimPrefix(versionInfo.Version, "v"), nil
}

// isVersionCompatible checks if version >= minVersion using semantic versioning
func isVersionCompatible(version, minVersion string) bool {
	v1Parts := parseVersion(version)
	v2Parts := parseVersion(minVersion)

	for i := 0; i < 3; i++ {
		if v1Parts[i] > v2Parts[i] {
			return true
		}
		if v1Parts[i] < v2Parts[i] {
			return false
		}
	}
	return true // versions are equal
}

// parseVersion parses a semantic version string into [major, minor, patch]
func parseVersion(version string) [3]int {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		fmt.Sscanf(parts[i], "%d", &result[i])
	}
	return result
}

// IsCLIAvailable returns true if the warpgate CLI is detected and available
func (w *WarpgateClient) IsCLIAvailable() bool {
	return w.cliDetected
}

// GetCLIVersion returns the detected warpgate CLI version
func (w *WarpgateClient) GetCLIVersion() string {
	return w.cliVersion
}

// GetBinaryPath returns the path to the warpgate binary
func (w *WarpgateClient) GetBinaryPath() string {
	return w.binaryPath
}

// GetRepoPath returns the repository path
func (w *WarpgateClient) GetRepoPath() string {
	return w.repoPath
}

// ExecuteTask runs a task command in the warpgate repository
func (w *WarpgateClient) ExecuteTask(taskName string, args map[string]string) (string, error) {
	// Build the task command
	cmdArgs := []string{taskName}

	// Add arguments in the format expected by taskfile
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "--")
		for key, value := range args {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%s=%s", key, value))
		}
	}

	cmd := exec.Command("task", cmdArgs...)
	cmd.Dir = w.repoPath
	cmd.Env = append(os.Environ(), "TASK_X_REMOTE_TASKFILES=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("task execution failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ListTasks returns available tasks from the Taskfile
func (w *WarpgateClient) ListTasks() ([]string, error) {
	cmd := exec.Command("task", "--list-all")
	cmd.Dir = w.repoPath
	cmd.Env = append(os.Environ(), "TASK_X_REMOTE_TASKFILES=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Parse the output to extract task names
	lines := strings.Split(string(output), "\n")
	var tasks []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and headers
		if line == "" || strings.HasPrefix(line, "task:") {
			continue
		}

		// Extract task name (first word)
		fields := strings.Fields(line)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "*") {
			tasks = append(tasks, fields[0])
		}
	}

	return tasks, nil
}

// ExecuteCLI runs a warpgate CLI command with the given arguments
func (w *WarpgateClient) ExecuteCLI(args ...string) (string, error) {
	if !w.cliDetected {
		return "", fmt.Errorf("warpgate CLI is not available")
	}

	cmd := exec.Command(w.binaryPath, args...)
	if w.repoPath != "" {
		cmd.Dir = w.repoPath
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("warpgate command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ExecuteCLIWithWorkdir runs a warpgate CLI command with explicit working directory
func (w *WarpgateClient) ExecuteCLIWithWorkdir(workdir string, args ...string) (string, error) {
	if !w.cliDetected {
		return "", fmt.Errorf("warpgate CLI is not available")
	}

	cmd := exec.Command(w.binaryPath, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("warpgate command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// WarpgateBuild executes the warpgate build command
func (w *WarpgateClient) WarpgateBuild(template string, opts BuildOptions) (string, error) {
	args := []string{"build"}

	if opts.Template != "" {
		args = append(args, "--template", opts.Template)
	} else if template != "" {
		args = append(args, template)
	}

	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	for _, arch := range opts.Architectures {
		args = append(args, "--arch", arch)
	}

	if opts.Push {
		args = append(args, "--push")
	}

	if opts.Registry != "" {
		args = append(args, "--registry", opts.Registry)
	}

	for key, value := range opts.Vars {
		args = append(args, "--var", fmt.Sprintf("%s=%s", key, value))
	}

	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	if opts.NoCache {
		args = append(args, "--no-cache")
	}

	if opts.SaveDigests {
		args = append(args, "--save-digests")
		if opts.DigestDir != "" {
			args = append(args, "--digest-dir", opts.DigestDir)
		}
	}

	return w.ExecuteCLI(args...)
}

// BuildOptions contains options for the warpgate build command
type BuildOptions struct {
	Template      string
	Target        string   // container, ami
	Architectures []string // amd64, arm64
	Push          bool
	Registry      string
	Vars          map[string]string
	Tags          []string
	NoCache       bool
	SaveDigests   bool
	DigestDir     string
}

// WarpgateValidate executes the warpgate validate command
func (w *WarpgateClient) WarpgateValidate(configPath string, syntaxOnly bool) (string, error) {
	args := []string{"validate"}

	if configPath != "" {
		args = append(args, configPath)
	}

	if syntaxOnly {
		args = append(args, "--syntax-only")
	}

	return w.ExecuteCLI(args...)
}

// WarpgateInit executes the warpgate init command
func (w *WarpgateClient) WarpgateInit(name string, opts InitOptions) (string, error) {
	args := []string{"init"}

	if name != "" {
		args = append(args, name)
	}

	if opts.OutputDir != "" {
		args = append(args, "--output", opts.OutputDir)
	}

	if opts.FromTemplate != "" {
		args = append(args, "--from", opts.FromTemplate)
	}

	return w.ExecuteCLI(args...)
}

// InitOptions contains options for the warpgate init command
type InitOptions struct {
	OutputDir    string
	FromTemplate string
}

// WarpgateTemplatesList lists templates from the registry
func (w *WarpgateClient) WarpgateTemplatesList(source, format string) (string, error) {
	args := []string{"templates", "list"}

	if source != "" {
		args = append(args, "--source", source)
	}

	if format != "" {
		args = append(args, "--format", format)
	}

	return w.ExecuteCLI(args...)
}

// WarpgateTemplatesInfo gets information about a template
func (w *WarpgateClient) WarpgateTemplatesInfo(template string) (string, error) {
	args := []string{"templates", "info", template}
	return w.ExecuteCLI(args...)
}

// WarpgateTemplatesAdd adds a template source
func (w *WarpgateClient) WarpgateTemplatesAdd(source string, name string) (string, error) {
	args := []string{"templates", "add"}

	if name != "" {
		args = append(args, name)
	}
	args = append(args, source)

	return w.ExecuteCLI(args...)
}

// WarpgateTemplatesRemove removes a template source
func (w *WarpgateClient) WarpgateTemplatesRemove(nameOrPath string) (string, error) {
	args := []string{"templates", "remove", nameOrPath}
	return w.ExecuteCLI(args...)
}

// WarpgateManifestsCreate creates a multi-arch manifest
func (w *WarpgateClient) WarpgateManifestsCreate(name string, images []string, push bool) (string, error) {
	args := []string{"manifests", "create", name}
	args = append(args, images...)

	if push {
		args = append(args, "--push")
	}

	return w.ExecuteCLI(args...)
}

// WarpgateManifestsPush pushes a manifest
func (w *WarpgateClient) WarpgateManifestsPush(name string, purge bool) (string, error) {
	args := []string{"manifests", "push", name}

	if purge {
		args = append(args, "--purge")
	}

	return w.ExecuteCLI(args...)
}

// WarpgateConfigGet gets a configuration value
func (w *WarpgateClient) WarpgateConfigGet(key string) (string, error) {
	args := []string{"config", "get"}

	if key != "" {
		args = append(args, key)
	}

	return w.ExecuteCLI(args...)
}

// WarpgateConfigSet sets a configuration value
func (w *WarpgateClient) WarpgateConfigSet(key, value string) (string, error) {
	args := []string{"config", "set", key, value}
	return w.ExecuteCLI(args...)
}

// WarpgateConfigShow shows current configuration
func (w *WarpgateClient) WarpgateConfigShow() (string, error) {
	args := []string{"config", "show"}
	return w.ExecuteCLI(args...)
}

// WarpgateConvert converts a Packer template to warpgate format
func (w *WarpgateClient) WarpgateConvert(source, output string) (string, error) {
	args := []string{"convert", "packer", source}

	if output != "" {
		args = append(args, "--output", output)
	}

	return w.ExecuteCLI(args...)
}

// WarpgateManifestsList lists available manifest tags for an image
func (w *WarpgateClient) WarpgateManifestsList(opts ManifestsListOptions) (string, error) {
	args := []string{"manifests", "list"}

	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}

	if opts.Registry != "" {
		args = append(args, "--registry", opts.Registry)
	}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	if opts.AuthFile != "" {
		args = append(args, "--auth-file", opts.AuthFile)
	}

	return w.ExecuteCLI(args...)
}

// ManifestsListOptions contains options for the manifests list command
type ManifestsListOptions struct {
	Name      string
	Registry  string
	Namespace string
	AuthFile  string
}

// WarpgateManifestsInspect inspects a multi-architecture manifest
func (w *WarpgateClient) WarpgateManifestsInspect(opts ManifestsInspectOptions) (string, error) {
	args := []string{"manifests", "inspect"}

	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}

	if opts.Registry != "" {
		args = append(args, "--registry", opts.Registry)
	}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	if opts.AuthFile != "" {
		args = append(args, "--auth-file", opts.AuthFile)
	}

	return w.ExecuteCLI(args...)
}

// ManifestsInspectOptions contains options for the manifests inspect command
type ManifestsInspectOptions struct {
	Name      string
	Registry  string
	Namespace string
	Tags      []string
	AuthFile  string
}

// WarpgateValidateConfig validates a warpgate config file
func (w *WarpgateClient) WarpgateValidateConfig(configPath string) (string, error) {
	args := []string{"validate"}

	if configPath != "" {
		args = append(args, configPath)
	}

	return w.ExecuteCLI(args...)
}
