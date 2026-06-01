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

// MinimumWarpgateVersion is the minimum supported warpgate CLI version.
// The CLI's Go module path is .../warpgate/v3, and several MCP tool contracts
// depend on v3-era flags (e.g. manifests create --name/--registry).
const MinimumWarpgateVersion = "3.0.0"

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

// WarpgateBuild executes the warpgate build command.
func (w *WarpgateClient) WarpgateBuild(ctx context.Context, template string, opts BuildOptions) (string, error) {
	return w.ExecuteCLI(ctx, buildArgs(template, opts)...)
}

// BuildOptions contains options for the warpgate build command.
//
// Fields mirror flags accepted by `warpgate build`. Many are target-specific
// (AMI/Azure/Proxmox) and ignored by other targets; the CLI is responsible
// for rejecting incompatible combinations.
type BuildOptions struct {
	Template      string
	FromGit       string // load template from git URL
	Target        string // container, ami, azure, proxmox
	Architectures []string
	Push          bool
	PushDigest    bool // push by digest, no tag (mutually exclusive with Push)
	Registry      string
	Vars          map[string]string
	VarFiles      []string // YAML var files
	BuildArgs     []string // build args (key=value)
	Tags          []string
	NoCache       bool
	SaveDigests   bool
	DigestDir     string

	// AMI / AWS
	Region          string
	InstanceType    string
	Force           bool
	Cleanup         bool
	DryRun          bool
	Regions         []string
	ParallelRegions bool
	CopyToRegions   []string
	StreamLogs      bool
	ShowEC2Status   bool
	OutputManifest  string

	// Azure
	AzureSubscription string
	AzureLocation     string
	AzureResourceGrp  string
	AzureGallery      string
	AzureImageDef     string
	AzureVMSize       string
	AzureIdentityID   string
	AzureTargetRegion []string
	AzureSubnetID     string
	AzureProxyVMSize  string

	// Proxmox
	ProxmoxEndpoint string
	ProxmoxNode     string
	ProxmoxStorage  string
	ProxmoxPool     string
}

// buildArgs converts BuildOptions to the argv `warpgate build` expects. Kept
// private and used by both WarpgateBuild and WarpgateBuildStreaming so the two
// can't drift.
func buildArgs(template string, opts BuildOptions) []string {
	args := []string{"build"}

	if opts.Template != "" {
		args = append(args, "--template", opts.Template)
	} else if template != "" {
		args = append(args, template)
	}
	if opts.FromGit != "" {
		args = append(args, "--from-git", opts.FromGit)
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
	if opts.PushDigest {
		args = append(args, "--push-digest")
	}
	if opts.Registry != "" {
		args = append(args, "--registry", opts.Registry)
	}
	for key, value := range opts.Vars {
		args = append(args, "--var", fmt.Sprintf("%s=%s", key, value))
	}
	for _, vf := range opts.VarFiles {
		args = append(args, "--var-file", vf)
	}
	for _, ba := range opts.BuildArgs {
		args = append(args, "--build-arg", ba)
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}
	if opts.NoCache {
		args = append(args, "--no-cache")
	}
	if opts.SaveDigests {
		args = append(args, "--save-digests")
	}
	if opts.DigestDir != "" {
		args = append(args, "--digest-dir", opts.DigestDir)
	}

	// AWS / AMI
	if opts.Region != "" {
		args = append(args, "--region", opts.Region)
	}
	if opts.InstanceType != "" {
		args = append(args, "--instance-type", opts.InstanceType)
	}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.Cleanup {
		args = append(args, "--cleanup")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	for _, r := range opts.Regions {
		args = append(args, "--regions", r)
	}
	if opts.ParallelRegions {
		args = append(args, "--parallel-regions")
	}
	for _, r := range opts.CopyToRegions {
		args = append(args, "--copy-to-regions", r)
	}
	if opts.StreamLogs {
		args = append(args, "--stream-logs")
	}
	if opts.ShowEC2Status {
		args = append(args, "--show-ec2-status")
	}
	if opts.OutputManifest != "" {
		args = append(args, "--output-manifest", opts.OutputManifest)
	}

	// Azure
	if opts.AzureSubscription != "" {
		args = append(args, "--subscription", opts.AzureSubscription)
	}
	if opts.AzureLocation != "" {
		args = append(args, "--location", opts.AzureLocation)
	}
	if opts.AzureResourceGrp != "" {
		args = append(args, "--resource-group", opts.AzureResourceGrp)
	}
	if opts.AzureGallery != "" {
		args = append(args, "--gallery", opts.AzureGallery)
	}
	if opts.AzureImageDef != "" {
		args = append(args, "--image-definition", opts.AzureImageDef)
	}
	if opts.AzureVMSize != "" {
		args = append(args, "--vm-size", opts.AzureVMSize)
	}
	if opts.AzureIdentityID != "" {
		args = append(args, "--identity-id", opts.AzureIdentityID)
	}
	for _, r := range opts.AzureTargetRegion {
		args = append(args, "--target-regions", r)
	}
	if opts.AzureSubnetID != "" {
		args = append(args, "--subnet-id", opts.AzureSubnetID)
	}
	if opts.AzureProxyVMSize != "" {
		args = append(args, "--proxy-vm-size", opts.AzureProxyVMSize)
	}

	// Proxmox
	if opts.ProxmoxEndpoint != "" {
		args = append(args, "--proxmox-endpoint", opts.ProxmoxEndpoint)
	}
	if opts.ProxmoxNode != "" {
		args = append(args, "--proxmox-node", opts.ProxmoxNode)
	}
	if opts.ProxmoxStorage != "" {
		args = append(args, "--proxmox-storage", opts.ProxmoxStorage)
	}
	if opts.ProxmoxPool != "" {
		args = append(args, "--proxmox-pool", opts.ProxmoxPool)
	}

	return args
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
func (w *WarpgateClient) WarpgateTemplatesList(ctx context.Context, source, format string, quiet bool) (string, error) {
	args := []string{"templates", "list"}

	if source != "" {
		args = append(args, "--source", source)
	}

	if format != "" {
		args = append(args, "--format", format)
	}

	if quiet {
		args = append(args, "--quiet")
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

// ManifestsCreateOptions contains options for `warpgate manifests create`.
// The current CLI takes no positional args; all inputs are flags. Registry
// is required by the CLI's PersistentPreRunE check.
type ManifestsCreateOptions struct {
	Name      string
	Registry  string
	Namespace string
	AuthFile  string
	Tags      []string
	DigestDir string
	DryRun    bool
	Force     bool

	// Verification & validation
	VerifyRegistry    *bool // pointer because CLI default is true (use --verify-registry=false to disable)
	VerifyConcurrency int
	MaxAge            string // e.g. "1h", "30m"

	// Architecture filtering
	RequireArch []string
	BestEffort  bool

	// OCI metadata
	Annotations []string // key=value
	Labels      []string // key=value

	// Behavior
	HealthCheck bool
	ShowDiff    bool
	NoProgress  bool
	Quiet       bool
	Verbose     bool
}

// WarpgateManifestsCreate creates and pushes a multi-arch manifest from
// digest files produced by prior `warpgate build --save-digests` invocations.
func (w *WarpgateClient) WarpgateManifestsCreate(ctx context.Context, opts ManifestsCreateOptions) (string, error) {
	args := []string{"manifests", "create"}

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
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}
	if opts.DigestDir != "" {
		args = append(args, "--digest-dir", opts.DigestDir)
	}
	if opts.VerifyRegistry != nil {
		args = append(args, fmt.Sprintf("--verify-registry=%t", *opts.VerifyRegistry))
	}
	if opts.VerifyConcurrency > 0 {
		args = append(args, "--verify-concurrency", fmt.Sprintf("%d", opts.VerifyConcurrency))
	}
	if opts.MaxAge != "" {
		args = append(args, "--max-age", opts.MaxAge)
	}
	for _, a := range opts.RequireArch {
		args = append(args, "--require-arch", a)
	}
	if opts.BestEffort {
		args = append(args, "--best-effort")
	}
	for _, a := range opts.Annotations {
		args = append(args, "--annotation", a)
	}
	for _, l := range opts.Labels {
		args = append(args, "--label", l)
	}
	if opts.HealthCheck {
		args = append(args, "--health-check")
	}
	if opts.ShowDiff {
		args = append(args, "--show-diff")
	}
	if opts.NoProgress {
		args = append(args, "--no-progress")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.Quiet {
		args = append(args, "--quiet")
	}
	if opts.Verbose {
		args = append(args, "--verbose")
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

// ConvertOptions contains options for `warpgate convert packer`.
type ConvertOptions struct {
	Source     string
	Output     string
	Author     string
	License    string
	Version    string
	BaseImage  string
	IncludeAMI *bool // pointer because CLI default is true (use --include-ami=false to disable)
	DryRun     bool
}

// WarpgateConvert converts a Packer template to warpgate format
func (w *WarpgateClient) WarpgateConvert(ctx context.Context, opts ConvertOptions) (string, error) {
	args := []string{"convert", "packer"}
	if opts.Source != "" {
		args = append(args, opts.Source)
	}
	if opts.Output != "" {
		args = append(args, "--output", opts.Output)
	}
	if opts.Author != "" {
		args = append(args, "--author", opts.Author)
	}
	if opts.License != "" {
		args = append(args, "--license", opts.License)
	}
	if opts.Version != "" {
		args = append(args, "--version", opts.Version)
	}
	if opts.BaseImage != "" {
		args = append(args, "--base-image", opts.BaseImage)
	}
	if opts.IncludeAMI != nil {
		args = append(args, fmt.Sprintf("--include-ami=%t", *opts.IncludeAMI))
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
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
	return w.ExecuteCLIStreaming(ctx, callback, buildArgs(template, opts)...)
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
			fmt.Fprintf(&results, "[DRY RUN] Would delete: %s\n", imageRef)
			continue
		}

		cmd := exec.CommandContext(ctx, toolPath, args...) //nolint:gosec // G204: registry tool execution with detected binary
		output, err := cmd.CombinedOutput()
		if err != nil {
			return results.String(), fmt.Errorf("failed to delete %s: %w\nOutput: %s", imageRef, err, string(output))
		}
		fmt.Fprintf(&results, "Deleted: %s\n", imageRef)
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
