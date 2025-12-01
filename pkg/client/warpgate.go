// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WarpgateClient handles interaction with the Warpgate repository
type WarpgateClient struct {
	repoPath string
}

// NewWarpgateClient creates a new Warpgate client
// If repoPath is empty, it will attempt to find the warpgate repo
func NewWarpgateClient(repoPath string) (*WarpgateClient, error) {
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

	return &WarpgateClient{
		repoPath: repoPath,
	}, nil
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

// ListTemplates returns available Packer templates
func (w *WarpgateClient) ListTemplates() ([]string, error) {
	templatesDir := filepath.Join(w.repoPath, "packer-templates")

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it contains Packer files
			templatePath := filepath.Join(templatesDir, entry.Name())
			if hasPackerFiles(templatePath) {
				templates = append(templates, entry.Name())
			}
		}
	}

	return templates, nil
}

// hasPackerFiles checks if a directory contains Packer template files
func hasPackerFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pkr.hcl") {
			return true
		}
	}

	return false
}

// GetTemplateInfo returns information about a specific template
func (w *WarpgateClient) GetTemplateInfo(templateName string) (map[string]interface{}, error) {
	templatePath := filepath.Join(w.repoPath, "packer-templates", templateName)

	if _, err := os.Stat(templatePath); err != nil {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	info := make(map[string]interface{})
	info["name"] = templateName
	info["path"] = templatePath

	// List files in the template directory
	entries, err := os.ReadDir(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	info["files"] = files

	// Check for README
	readmePath := filepath.Join(templatePath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		info["readme"] = string(content)
	}

	return info, nil
}
