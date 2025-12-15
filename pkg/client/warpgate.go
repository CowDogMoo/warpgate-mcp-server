// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WarpgateClient handles interaction with the Warpgate CLI
type WarpgateClient struct {
	repoPath string
}

// TemplateInfo represents information about a warpgate template
type TemplateInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Source      string   `json:"source,omitempty"`
	Path        string   `json:"path,omitempty"`
}

// WarpgateConfig represents the warpgate configuration
type WarpgateConfig struct {
	Registry struct {
		Default string `yaml:"default"`
	} `yaml:"registry"`
	Templates struct {
		CacheDir     string            `yaml:"cache_dir"`
		Repositories map[string]string `yaml:"repositories"`
		LocalPaths   []string          `yaml:"local_paths"`
	} `yaml:"templates"`
	Build struct {
		DefaultArch []string `yaml:"default_arch"`
	} `yaml:"build"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
}

// NewWarpgateClient creates a new Warpgate client
// If repoPath is empty, it will attempt to find the warpgate repo
func NewWarpgateClient(repoPath string) (*WarpgateClient, error) {
	// First verify that warpgate CLI is available
	if _, err := exec.LookPath("warpgate"); err != nil {
		return nil, fmt.Errorf("warpgate CLI not found in PATH. Please install warpgate: https://github.com/cowdogmoo/warpgate")
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
			// If we can't find the repo, that's okay - warpgate can work without it
			// We'll just use the current directory
			if cwd, err := os.Getwd(); err == nil {
				repoPath = cwd
			}
		}
	}

	// Verify the path exists
	if _, err := os.Stat(repoPath); err != nil {
		return nil, fmt.Errorf("path does not exist: %s", repoPath)
	}

	return &WarpgateClient{
		repoPath: repoPath,
	}, nil
}

// GetRepoPath returns the repository path
func (w *WarpgateClient) GetRepoPath() string {
	return w.repoPath
}

// ExecuteWarpgate runs a warpgate CLI command and returns the output
func (w *WarpgateClient) ExecuteWarpgate(args ...string) (string, error) {
	cmd := exec.Command("warpgate", args...)
	cmd.Dir = w.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("warpgate command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ExecuteWarpgateJSON runs a warpgate CLI command and parses JSON output
func (w *WarpgateClient) ExecuteWarpgateJSON(result interface{}, args ...string) error {
	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(output), result); err != nil {
		return fmt.Errorf("failed to parse JSON output: %w\nOutput: %s", err, output)
	}

	return nil
}

// GetWarpgateConfig loads the warpgate configuration
func (w *WarpgateClient) GetWarpgateConfig() (*WarpgateConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Try XDG config location first
	configPaths := []string{
		filepath.Join(home, ".config", "warpgate", "config.yaml"),
		filepath.Join(home, ".warpgate", "config.yaml"),
		filepath.Join(w.repoPath, "config.yaml"),
	}

	var configData []byte
	var foundPath string

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			configData = data
			foundPath = path
			break
		}
	}

	if configData == nil {
		return nil, fmt.Errorf("warpgate config not found in any of: %v", configPaths)
	}

	var config WarpgateConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config at %s: %w", foundPath, err)
	}

	return &config, nil
}

// ListTemplates returns available warpgate templates from all configured sources
func (w *WarpgateClient) ListTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Use warpgate templates list --format json
	if err := w.ExecuteWarpgateJSON(&templates, "templates", "list", "--format", "json", "--quiet"); err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// ListTemplatesFromSource returns templates from a specific source
func (w *WarpgateClient) ListTemplatesFromSource(source string) ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Use warpgate templates list --format json --source <source>
	if err := w.ExecuteWarpgateJSON(&templates, "templates", "list", "--format", "json", "--source", source, "--quiet"); err != nil {
		return nil, fmt.Errorf("failed to list templates from source %s: %w", source, err)
	}

	return templates, nil
}

// SearchTemplates searches for templates with fuzzy matching
func (w *WarpgateClient) SearchTemplates(query string) ([]TemplateInfo, error) {
	// Use warpgate templates search command
	_, err := w.ExecuteWarpgate("templates", "search", query)
	if err != nil {
		return nil, fmt.Errorf("failed to search templates: %w", err)
	}

	// Parse the output (warpgate search outputs text format)
	// For now, we'll fall back to listing and filtering
	allTemplates, err := w.ListTemplates()
	if err != nil {
		return nil, err
	}

	// Simple filter by query in name, description, or tags
	var results []TemplateInfo
	queryLower := strings.ToLower(query)
	for _, t := range allTemplates {
		if strings.Contains(strings.ToLower(t.Name), queryLower) ||
			strings.Contains(strings.ToLower(t.Description), queryLower) {
			results = append(results, t)
			continue
		}
		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, t)
				break
			}
		}
	}

	return results, nil
}

// GetTemplateInfo returns detailed information about a specific template
func (w *WarpgateClient) GetTemplateInfo(templateName string) (map[string]interface{}, error) {
	// Use warpgate templates info command
	output, err := w.ExecuteWarpgate("templates", "info", templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to get template info: %w", err)
	}

	// Parse the text output and convert to map
	info := make(map[string]interface{})
	info["name"] = templateName
	info["output"] = output

	// Try to extract key information from the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Description:") {
			info["description"] = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		} else if strings.HasPrefix(line, "Version:") {
			info["version"] = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Author:") {
			info["author"] = strings.TrimSpace(strings.TrimPrefix(line, "Author:"))
		}
	}

	return info, nil
}

// ValidateTemplate validates a warpgate template
func (w *WarpgateClient) ValidateTemplate(templatePath string, syntaxOnly bool) (string, error) {
	args := []string{"validate", templatePath}
	if syntaxOnly {
		args = append(args, "--syntax-only")
	}

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// BuildTemplate builds a warpgate template
func (w *WarpgateClient) BuildTemplate(opts BuildOptions) (string, error) {
	args := []string{"build"}

	// Add template name or path
	if opts.Template != "" {
		args = append(args, "--template", opts.Template)
	} else if opts.Config != "" {
		args = append(args, opts.Config)
	}

	// Add target override
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	// Add architectures
	if len(opts.Architectures) > 0 {
		args = append(args, "--arch", strings.Join(opts.Architectures, ","))
	}

	// Add push flag
	if opts.Push {
		args = append(args, "--push")
	}

	// Add registry override
	if opts.Registry != "" {
		args = append(args, "--registry", opts.Registry)
	}

	// Add tags
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	// Add variables
	for key, value := range opts.Variables {
		args = append(args, "--var", fmt.Sprintf("%s=%s", key, value))
	}

	// Add no-cache flag
	if opts.NoCache {
		args = append(args, "--no-cache")
	}

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// BuildOptions represents options for building a template
type BuildOptions struct {
	Template      string
	Config        string
	Target        string
	Architectures []string
	Push          bool
	Registry      string
	Tags          []string
	Variables     map[string]string
	NoCache       bool
}

// InitTemplate initializes a new warpgate template
func (w *WarpgateClient) InitTemplate(name, fromTemplate, outputDir string) (string, error) {
	args := []string{"init", name}

	if fromTemplate != "" {
		args = append(args, "--from", fromTemplate)
	}

	if outputDir != "" {
		args = append(args, "--output", outputDir)
	}

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// AddTemplateSource adds a template source (git repository or local path)
func (w *WarpgateClient) AddTemplateSource(urlOrPath, name string) (string, error) {
	args := []string{"templates", "add"}

	if name != "" {
		args = append(args, name)
	}

	args = append(args, urlOrPath)

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// RemoveTemplateSource removes a template source
func (w *WarpgateClient) RemoveTemplateSource(pathOrName string) (string, error) {
	output, err := w.ExecuteWarpgate("templates", "remove", pathOrName)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// UpdateTemplateCache updates the template cache
func (w *WarpgateClient) UpdateTemplateCache() (string, error) {
	output, err := w.ExecuteWarpgate("templates", "update")
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// ConvertPackerTemplate converts a Packer template to warpgate format
func (w *WarpgateClient) ConvertPackerTemplate(opts ConvertOptions) (string, error) {
	args := []string{"convert", "packer", opts.TemplateDir}

	if opts.Output != "" {
		args = append(args, "--output", opts.Output)
	}

	if opts.Author != "" {
		args = append(args, "--author", opts.Author)
	}

	if opts.Version != "" {
		args = append(args, "--version", opts.Version)
	}

	if !opts.IncludeAMI {
		args = append(args, "--include-ami=false")
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// ConvertOptions represents options for converting Packer templates
type ConvertOptions struct {
	TemplateDir string
	Output      string
	Author      string
	Version     string
	IncludeAMI  bool
	DryRun      bool
}

// CreateManifest creates a multi-architecture image manifest
func (w *WarpgateClient) CreateManifest(opts ManifestOptions) (string, error) {
	args := []string{"manifests", "create"}

	args = append(args, "--name", opts.ImageName)
	args = append(args, "--registry", opts.Registry)

	for _, tag := range opts.Tags {
		args = append(args, "--tags", tag)
	}

	if opts.DigestDir != "" {
		args = append(args, "--digest-dir", opts.DigestDir)
	}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	for _, arch := range opts.RequiredArchitectures {
		args = append(args, "--require-arch", arch)
	}

	output, err := w.ExecuteWarpgate(args...)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// ManifestOptions represents options for creating manifests
type ManifestOptions struct {
	ImageName             string
	Registry              string
	Tags                  []string
	DigestDir             string
	Namespace             string
	RequiredArchitectures []string
}

// ExecuteTask runs a task command in the warpgate repository
// This is kept for backward compatibility with Taskfile operations
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
// This is kept for backward compatibility with Taskfile operations
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
