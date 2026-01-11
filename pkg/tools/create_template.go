// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type templateData struct {
	TemplateName     string
	Description      string
	BaseImage        string
	BaseImageVersion string
	IncludeAMI       bool
	Title            string
	Platforms        []string
}

func createTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "create_template",
		Description: "Create a new warpgate template with warpgate.yaml configuration and scaffolding. Generates the modern YAML-based template format.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the new template (e.g., 'my-awesome-template')",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Brief description of what this template creates",
				},
				"base_image": map[string]interface{}{
					"type":        "string",
					"description": "Base Docker image to use (default: 'ubuntu')",
				},
				"base_image_version": map[string]interface{}{
					"type":        "string",
					"description": "Version of the base image (default: '22.04')",
				},
				"platforms": map[string]interface{}{
					"type":        "array",
					"description": "Target platforms (default: ['linux/amd64', 'linux/arm64'])",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AWS AMI target configuration (default: false)",
				},
			},
			Required: []string{"template_name", "description"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		templateName := args["template_name"].(string)
		description := args["description"].(string)

		// Validate template name
		if !isValidTemplateName(templateName) {
			return mcp.NewToolResultError(
				"Invalid template name. Use lowercase letters, numbers, and hyphens only."), nil
		}

		// Set defaults
		baseImage := "ubuntu"
		if val, ok := args["base_image"].(string); ok && val != "" {
			baseImage = val
		}

		baseImageVersion := "22.04"
		if val, ok := args["base_image_version"].(string); ok && val != "" {
			baseImageVersion = val
		}

		includeAMI := false
		if val, ok := args["include_ami"].(bool); ok {
			includeAMI = val
		}

		var platforms []string
		if val, ok := args["platforms"].([]interface{}); ok {
			for _, p := range val {
				if ps, ok := p.(string); ok {
					platforms = append(platforms, ps)
				}
			}
		}

		// Create template directory
		templateDir := filepath.Join(warpgatePath, "templates", templateName)
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			logger.Errorf("Failed to create template directory: %v", err)
			return mcp.NewToolResultError(
				fmt.Sprintf("Failed to create template directory: %v", err)), nil
		}

		// Check if template already exists
		if _, err := os.Stat(filepath.Join(templateDir, "warpgate.yaml")); err == nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("Template '%s' already exists at %s", templateName, templateDir)), nil
		}

		// Create scripts directory
		scriptsDir := filepath.Join(templateDir, "scripts")
		if err := os.MkdirAll(scriptsDir, 0755); err != nil {
			logger.Errorf("Failed to create scripts directory: %v", err)
			return mcp.NewToolResultError(
				fmt.Sprintf("Failed to create scripts directory: %v", err)), nil
		}

		// Prepare template data
		title := strings.ReplaceAll(strings.Title(strings.ReplaceAll(templateName, "-", " ")), " ", "-")
		data := templateData{
			TemplateName:     templateName,
			Description:      description,
			BaseImage:        baseImage,
			BaseImageVersion: baseImageVersion,
			IncludeAMI:       includeAMI,
			Title:            title,
			Platforms:        platforms,
		}

		// Create files from templates
		files := map[string]string{
			"warpgate.yaml": "templates/warpgate.yaml.tmpl",
			"README.md":     "templates/warpgate-README.md.tmpl",
		}

		for filename, templatePath := range files {
			filePath := filepath.Join(templateDir, filename)
			if err := renderTemplate(templatePath, filePath, data); err != nil {
				logger.Errorf("Failed to create %s: %v", filename, err)
				return mcp.NewToolResultError(
					fmt.Sprintf("Failed to create %s: %v", filename, err)), nil
			}
		}

		// Create a placeholder setup script
		setupScriptPath := filepath.Join(scriptsDir, "setup.sh")
		setupScriptContent := `#!/bin/bash
# Provisioning script for ` + templateName + `
# Add your setup commands here

set -e

echo "Setting up ` + templateName + `..."

# Example: Install packages
# apt-get update
# apt-get install -y your-packages

echo "Setup complete!"
`
		if err := os.WriteFile(setupScriptPath, []byte(setupScriptContent), 0755); err != nil {
			logger.Errorf("Failed to create setup script: %v", err)
			return mcp.NewToolResultError(
				fmt.Sprintf("Failed to create setup script: %v", err)), nil
		}

		logger.Infof("Successfully created template: %s", templateName)
		result := fmt.Sprintf(`Successfully created warpgate template '%s' at %s

Files created:
- warpgate.yaml (Template configuration)
- README.md (Template documentation)
- scripts/setup.sh (Provisioning script placeholder)
%s
Next steps:
1. Edit warpgate.yaml to configure provisioning
2. Add provisioning logic to scripts/setup.sh or use Ansible
3. Validate: warpgate validate %s/warpgate.yaml
4. Build: warpgate build %s/warpgate.yaml
`, templateName, templateDir,
			func() string {
				if includeAMI {
					return "- AMI target configuration included\n"
				}
				return ""
			}(), templateDir, templateDir)

		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}

func renderTemplate(templatePath, outputPath string, data templateData) error {
	// Read template file from embedded FS
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return nil
}

func isValidTemplateName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
