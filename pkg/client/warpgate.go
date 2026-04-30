// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
// repoPath is optional and used as working directory for CLI commands
func NewWarpgateClient(repoPath string) (*WarpgateClient, error) {
	client := &WarpgateClient{}

	// Detect warpgate CLI binary
	if err := client.detectWarpgateBinary(""); err != nil {
		client.cliDetected = false
	} else {
		client.cliDetected = true
	}

	if repoPath != "" {
		// Verify the path exists
		if _, err := os.Stat(repoPath); err != nil {
			return nil, fmt.Errorf("path not found: %s: %w", repoPath, err)
		}
		client.repoPath = repoPath
	}

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
		// Verify the path exists
		if _, err := os.Stat(repoPath); err != nil {
			return nil, fmt.Errorf("path not found: %s: %w", repoPath, err)
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

		outputStr := string(output)
		if strings.Contains(strings.ToLower(outputStr), "dev") {
			return "dev", nil
		}

		// Parse version from output like "warpgate version v3.0.1"
		re := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) < 2 {
			return "", fmt.Errorf("could not parse version from output: %s", outputStr)
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
	if strings.ToLower(version) == "dev" {
		return true
	}

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
		_, _ = fmt.Sscanf(parts[i], "%d", &result[i])
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

// ExecuteCLI runs a warpgate CLI command with the given arguments
func (w *WarpgateClient) ExecuteCLI(ctx context.Context, args ...string) (string, error) {
	if !w.cliDetected {
		return "", fmt.Errorf("warpgate CLI is not available")
	}

	cmd := exec.CommandContext(ctx, w.binaryPath, args...) //nolint:gosec // G204: warpgate CLI execution with validated binary
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
func (w *WarpgateClient) ExecuteCLIWithWorkdir(ctx context.Context, workdir string, args ...string) (string, error) {
	if !w.cliDetected {
		return "", fmt.Errorf("warpgate CLI is not available")
	}

	cmd := exec.CommandContext(ctx, w.binaryPath, args...) //nolint:gosec // G204: warpgate CLI execution with validated binary
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
func (w *WarpgateClient) WarpgateBuild(ctx context.Context, template string, opts BuildOptions) (string, error) {
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

	return w.ExecuteCLI(ctx, args...)
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
func (w *WarpgateClient) WarpgateValidate(ctx context.Context, configPath string, syntaxOnly bool) (string, error) {
	args := []string{"validate"}

	if configPath != "" {
		args = append(args, configPath)
	}

	if syntaxOnly {
		args = append(args, "--syntax-only")
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateInit executes the warpgate init command
func (w *WarpgateClient) WarpgateInit(ctx context.Context, name string, opts InitOptions) (string, error) {
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

	return w.ExecuteCLI(ctx, args...)
}

// InitOptions contains options for the warpgate init command
type InitOptions struct {
	OutputDir    string
	FromTemplate string
}

// WarpgateTemplatesList lists templates from the registry
func (w *WarpgateClient) WarpgateTemplatesList(ctx context.Context, source, format string) (string, error) {
	args := []string{"templates", "list"}

	if source != "" {
		args = append(args, "--source", source)
	}

	if format != "" {
		args = append(args, "--format", format)
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateTemplatesInfo gets information about a template
func (w *WarpgateClient) WarpgateTemplatesInfo(ctx context.Context, template string) (string, error) {
	args := []string{"templates", "info", template}
	return w.ExecuteCLI(ctx, args...)
}

// WarpgateTemplatesAdd adds a template source
func (w *WarpgateClient) WarpgateTemplatesAdd(ctx context.Context, source string, name string) (string, error) {
	args := []string{"templates", "add"}

	if name != "" {
		args = append(args, name)
	}
	args = append(args, source)

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateTemplatesRemove removes a template source
func (w *WarpgateClient) WarpgateTemplatesRemove(ctx context.Context, nameOrPath string) (string, error) {
	args := []string{"templates", "remove", nameOrPath}
	return w.ExecuteCLI(ctx, args...)
}

// WarpgateManifestsCreate creates a multi-arch manifest
func (w *WarpgateClient) WarpgateManifestsCreate(ctx context.Context, name string, images []string, push bool) (string, error) {
	args := []string{"manifests", "create", name}
	args = append(args, images...)

	if push {
		args = append(args, "--push")
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateManifestsPush pushes a manifest
func (w *WarpgateClient) WarpgateManifestsPush(ctx context.Context, name string, purge bool) (string, error) {
	args := []string{"manifests", "push", name}

	if purge {
		args = append(args, "--purge")
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateConfigGet gets a configuration value
func (w *WarpgateClient) WarpgateConfigGet(ctx context.Context, key string) (string, error) {
	args := []string{"config", "get"}

	if key != "" {
		args = append(args, key)
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateConfigSet sets a configuration value
func (w *WarpgateClient) WarpgateConfigSet(ctx context.Context, key, value string) (string, error) {
	args := []string{"config", "set", key, value}
	return w.ExecuteCLI(ctx, args...)
}

// WarpgateConfigShow shows current configuration
func (w *WarpgateClient) WarpgateConfigShow(ctx context.Context) (string, error) {
	args := []string{"config", "show"}
	return w.ExecuteCLI(ctx, args...)
}

// WarpgateConvert converts a Packer template to warpgate format
func (w *WarpgateClient) WarpgateConvert(ctx context.Context, source, output string) (string, error) {
	args := []string{"convert", "packer", source}

	if output != "" {
		args = append(args, "--output", output)
	}

	return w.ExecuteCLI(ctx, args...)
}

// WarpgateManifestsList lists available manifest tags for an image
func (w *WarpgateClient) WarpgateManifestsList(ctx context.Context, opts ManifestsListOptions) (string, error) {
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

	return w.ExecuteCLI(ctx, args...)
}

// ManifestsListOptions contains options for the manifests list command
type ManifestsListOptions struct {
	Name      string
	Registry  string
	Namespace string
	AuthFile  string
}

// WarpgateManifestsInspect inspects a multi-architecture manifest
func (w *WarpgateClient) WarpgateManifestsInspect(ctx context.Context, opts ManifestsInspectOptions) (string, error) {
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

	return w.ExecuteCLI(ctx, args...)
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
func (w *WarpgateClient) WarpgateValidateConfig(ctx context.Context, configPath string) (string, error) {
	args := []string{"validate"}

	if configPath != "" {
		args = append(args, configPath)
	}

	return w.ExecuteCLI(ctx, args...)
}

// OutputCallback is called for each line of output during streaming execution
type OutputCallback func(line string)

// splitLinesOrCarriageReturn is a custom split function for bufio.Scanner
// that splits on both newlines (\n) and carriage returns (\r)
func splitLinesOrCarriageReturn(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Look for \n or \r
	if i := strings.IndexAny(string(data), "\n\r"); i >= 0 {
		// Return the line (without the delimiter)
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, return remaining data
	if atEOF {
		return len(data), data, nil
	}

	// Request more data
	return 0, nil, nil
}

// ExecuteCLIStreaming runs a warpgate CLI command with streaming output
func (w *WarpgateClient) ExecuteCLIStreaming(ctx context.Context, callback OutputCallback, args ...string) (string, error) {
	if !w.cliDetected {
		return "", fmt.Errorf("warpgate CLI is not available")
	}

	cmd := exec.CommandContext(ctx, w.binaryPath, args...) //nolint:gosec // G204: warpgate CLI execution with validated binary
	if w.repoPath != "" {
		cmd.Dir = w.repoPath
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Collect all output
	var outputBuilder strings.Builder
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Read stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		scanner.Split(splitLinesOrCarriageReturn)
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			outputBuilder.WriteString(line)
			outputBuilder.WriteString("\n")
			mu.Unlock()
			if callback != nil {
				callback(line)
			}
		}
	}()

	// Read stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Split(splitLinesOrCarriageReturn)
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			outputBuilder.WriteString(line)
			outputBuilder.WriteString("\n")
			mu.Unlock()
			if callback != nil {
				callback(line)
			}
		}
	}()

	// Wait for goroutines to finish
	wg.Wait()

	// Wait for command to complete
	err = cmd.Wait()
	output := outputBuilder.String()

	if err != nil {
		return output, fmt.Errorf("warpgate command failed: %w\nOutput: %s", err, output)
	}

	return output, nil
}

// WarpgateBuildStreaming executes the warpgate build command with streaming output
func (w *WarpgateClient) WarpgateBuildStreaming(ctx context.Context, template string, opts BuildOptions, callback OutputCallback) (string, error) {
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

	return w.ExecuteCLIStreaming(ctx, callback, args...)
}

// RegistryDeleteOptions contains options for deleting images from a registry
type RegistryDeleteOptions struct {
	Name      string
	Registry  string
	Namespace string
	Tags      []string
	AuthFile  string
	DryRun    bool
}

// RegistryCopyOptions contains options for copying images between registries
type RegistryCopyOptions struct {
	SourceImage     string
	DestImage       string
	SourceAuth      string
	DestAuth        string
	AllTags         bool
	PreserveDigests bool
}

// DetectRegistryTool finds an available registry management tool
func DetectRegistryTool() (string, error) {
	tools := []string{"skopeo", "crane", "docker", "podman"}
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no registry tool found (tried: skopeo, crane, docker, podman)")
}

// RegistryDelete deletes an image from a container registry
func (w *WarpgateClient) RegistryDelete(ctx context.Context, opts RegistryDeleteOptions) (string, error) {
	toolPath, err := DetectRegistryTool()
	if err != nil {
		return "", err
	}

	tool := filepath.Base(toolPath)
	var results strings.Builder

	for _, tag := range opts.Tags {
		imageRef := buildImageRef(opts.Registry, opts.Namespace, opts.Name, tag)

		var args []string
		switch tool {
		case "skopeo":
			args = []string{"delete"}
			if opts.AuthFile != "" {
				args = append(args, "--authfile", opts.AuthFile)
			}
			args = append(args, "docker://"+imageRef)
		case "crane":
			args = []string{"delete"}
			args = append(args, imageRef)
		case "docker", "podman":
			// Docker/Podman use rmi for local, but we need registry delete
			// These tools don't support direct registry deletion well
			return "", fmt.Errorf("%s does not support direct registry deletion; use skopeo or crane", tool)
		}

		if opts.DryRun {
			results.WriteString(fmt.Sprintf("[DRY RUN] Would delete: %s\n", imageRef))
			continue
		}

		cmd := exec.CommandContext(ctx, toolPath, args...) //nolint:gosec // G204: registry tool execution with detected binary
		output, err := cmd.CombinedOutput()
		if err != nil {
			return results.String(), fmt.Errorf("failed to delete %s: %w\nOutput: %s", imageRef, err, string(output))
		}
		results.WriteString(fmt.Sprintf("Deleted: %s\n", imageRef))
	}

	return results.String(), nil
}

// RegistryCopy copies an image between registries
func (w *WarpgateClient) RegistryCopy(ctx context.Context, opts RegistryCopyOptions) (string, error) {
	toolPath, err := DetectRegistryTool()
	if err != nil {
		return "", err
	}

	tool := filepath.Base(toolPath)
	var args []string

	switch tool {
	case "skopeo":
		args = []string{"copy"}
		if opts.AllTags {
			args = append(args, "--all")
		}
		if opts.PreserveDigests {
			args = append(args, "--preserve-digests")
		}
		if opts.SourceAuth != "" {
			args = append(args, "--src-authfile", opts.SourceAuth)
		}
		if opts.DestAuth != "" {
			args = append(args, "--dest-authfile", opts.DestAuth)
		}
		args = append(args, "docker://"+opts.SourceImage, "docker://"+opts.DestImage)
	case "crane":
		if opts.AllTags {
			args = []string{"copy", "--all-tags"}
		} else {
			args = []string{"copy"}
		}
		args = append(args, opts.SourceImage, opts.DestImage)
	case "docker", "podman":
		// Docker/Podman require pull-tag-push workflow
		return "", fmt.Errorf("%s requires pull-tag-push workflow; use skopeo or crane for direct copy", tool)
	}

	cmd := exec.CommandContext(ctx, toolPath, args...) //nolint:gosec // G204: registry tool execution with detected binary
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to copy image: %w\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("Successfully copied %s to %s\n%s", opts.SourceImage, opts.DestImage, string(output)), nil
}

// buildImageRef constructs a full image reference from components
func buildImageRef(registry, namespace, name, tag string) string {
	var ref strings.Builder
	if registry != "" {
		ref.WriteString(registry)
		ref.WriteString("/")
	}
	if namespace != "" {
		ref.WriteString(namespace)
		ref.WriteString("/")
	}
	ref.WriteString(name)
	if tag != "" {
		ref.WriteString(":")
		ref.WriteString(tag)
	}
	return ref.String()
}
