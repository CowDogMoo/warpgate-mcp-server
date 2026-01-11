// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

// WarpgateTemplateConfig represents the structure of a warpgate.yaml file
type WarpgateTemplateConfig struct {
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	Version      string                 `yaml:"version"`
	Maintainers  []string               `yaml:"maintainers"`
	BaseImage    *BaseImageConfig       `yaml:"base_image"`
	Targets      *TargetsConfig         `yaml:"targets"`
	Provisioners []ProvisionerConfig    `yaml:"provisioners"`
	Variables    map[string]interface{} `yaml:"variables"`
}

// BaseImageConfig represents the base image configuration
type BaseImageConfig struct {
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`
}

// TargetsConfig represents build target configurations
type TargetsConfig struct {
	Container *ContainerTargetConfig `yaml:"container"`
	AMI       *AMITargetConfig       `yaml:"ami"`
}

// ContainerTargetConfig represents container build target settings
type ContainerTargetConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Architectures []string `yaml:"architectures"`
}

// AMITargetConfig represents AWS AMI build target settings
type AMITargetConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Region       string `yaml:"region"`
	InstanceType string `yaml:"instance_type"`
}

// ProvisionerConfig represents a provisioning step
type ProvisionerConfig struct {
	Type        string   `yaml:"type"`
	Name        string   `yaml:"name"`
	Inline      []string `yaml:"inline"`
	Script      string   `yaml:"script"`
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
}

func warpgateSchemaValidate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_schema_validate",
		Description: "Validate a warpgate.yaml configuration file against the template schema. Reports structural errors, missing required fields, and type mismatches.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"config_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the warpgate.yaml file to validate",
				},
				"template_dir": map[string]interface{}{
					"type":        "string",
					"description": "Path to a template directory containing warpgate.yaml",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var configPath string

		if path, ok := request.Params.Arguments["config_path"].(string); ok && path != "" {
			configPath = path
		} else if dir, ok := request.Params.Arguments["template_dir"].(string); ok && dir != "" {
			configPath = filepath.Join(dir, "warpgate.yaml")
		} else {
			return mcp.NewToolResultError("either config_path or template_dir is required"), nil
		}

		// Read and parse the YAML file
		data, err := os.ReadFile(configPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read config file: %v", err)), nil
		}

		var config WarpgateTemplateConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid YAML syntax: %v", err)), nil
		}

		// Validate required fields and structure
		var errors []string
		var warnings []string

		// Required: name
		if config.Name == "" {
			errors = append(errors, "missing required field: name")
		} else {
			// Validate name format
			if !isValidTemplateName(config.Name) {
				errors = append(errors, fmt.Sprintf("invalid name format: '%s' (must be lowercase alphanumeric with hyphens, no leading/trailing hyphens)", config.Name))
			}
		}

		// Recommended: description
		if config.Description == "" {
			warnings = append(warnings, "missing recommended field: description")
		}

		// Validate version format if present
		if config.Version != "" && !isValidVersion(config.Version) {
			warnings = append(warnings, fmt.Sprintf("version '%s' does not follow semantic versioning (e.g., v1.0.0)", config.Version))
		}

		// Validate targets if present
		if config.Targets != nil {
			if config.Targets.Container != nil {
				for _, arch := range config.Targets.Container.Architectures {
					if arch != "amd64" && arch != "arm64" {
						errors = append(errors, fmt.Sprintf("invalid architecture '%s' in container target (must be 'amd64' or 'arm64')", arch))
					}
				}
			}
		}

		// Validate provisioners if present
		for i, prov := range config.Provisioners {
			if prov.Type == "" {
				errors = append(errors, fmt.Sprintf("provisioner[%d] missing required field: type", i))
			} else if prov.Type != "shell" && prov.Type != "file" && prov.Type != "ansible" {
				warnings = append(warnings, fmt.Sprintf("provisioner[%d] has non-standard type: '%s'", i, prov.Type))
			}

			// Shell provisioners need inline or script
			if prov.Type == "shell" && len(prov.Inline) == 0 && prov.Script == "" {
				errors = append(errors, fmt.Sprintf("provisioner[%d] (shell) requires either 'inline' or 'script'", i))
			}

			// File provisioners need source and destination
			if prov.Type == "file" && (prov.Source == "" || prov.Destination == "") {
				errors = append(errors, fmt.Sprintf("provisioner[%d] (file) requires both 'source' and 'destination'", i))
			}
		}

		// Build result
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Validation results for: %s\n\n", configPath))

		if len(errors) == 0 && len(warnings) == 0 {
			result.WriteString("✓ Configuration is valid\n\n")
			result.WriteString(fmt.Sprintf("Template: %s\n", config.Name))
			if config.Description != "" {
				result.WriteString(fmt.Sprintf("Description: %s\n", config.Description))
			}
			if config.Version != "" {
				result.WriteString(fmt.Sprintf("Version: %s\n", config.Version))
			}
		} else {
			if len(errors) > 0 {
				result.WriteString(fmt.Sprintf("✗ Found %d error(s):\n", len(errors)))
				for _, err := range errors {
					result.WriteString(fmt.Sprintf("  - %s\n", err))
				}
				result.WriteString("\n")
			}

			if len(warnings) > 0 {
				result.WriteString(fmt.Sprintf("⚠ Found %d warning(s):\n", len(warnings)))
				for _, warn := range warnings {
					result.WriteString(fmt.Sprintf("  - %s\n", warn))
				}
			}
		}

		// Also run warpgate CLI validation if available
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err == nil && wg.IsCLIAvailable() {
			output, err := wg.WarpgateValidateConfig(configPath)
			if err != nil {
				result.WriteString(fmt.Sprintf("\nCLI validation failed: %v\n%s", err, output))
			} else {
				result.WriteString(fmt.Sprintf("\n\nCLI validation: %s", output))
			}
		}

		if len(errors) > 0 {
			return mcp.NewToolResultError(result.String()), nil
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	s.AddTool(tool, handler)
}

// isValidVersion checks if the version follows semantic versioning
func isValidVersion(version string) bool {
	v := strings.TrimPrefix(version, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, part := range parts {
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}
