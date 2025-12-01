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
}

func createTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "create_template",
		Description: "Create a new Packer template with all required files and structure",
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
					"description": "Version of the base image (default: 'latest')",
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AWS AMI configuration (default: false)",
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

		baseImageVersion := "latest"
		if val, ok := args["base_image_version"].(string); ok && val != "" {
			baseImageVersion = val
		}

		includeAMI := false
		if val, ok := args["include_ami"].(bool); ok {
			includeAMI = val
		}

		// Create template directory
		templateDir := filepath.Join(warpgatePath, "packer-templates", templateName)
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			logger.Errorf("Failed to create template directory: %v", err)
			return mcp.NewToolResultError(
				fmt.Sprintf("Failed to create template directory: %v", err)), nil
		}

		// Check if template already exists
		if _, err := os.Stat(filepath.Join(templateDir, "docker.pkr.hcl")); err == nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("Template '%s' already exists at %s", templateName, templateDir)), nil
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
		}

		// Create files from templates
		files := map[string]string{
			"plugins.pkr.hcl":   "templates/plugins.pkr.hcl.tmpl",
			"locals.pkr.hcl":    "templates/locals.pkr.hcl.tmpl",
			"variables.pkr.hcl": "templates/variables.pkr.hcl.tmpl",
			"docker.pkr.hcl":    "templates/docker.pkr.hcl.tmpl",
			"README.md":         "templates/README.md.tmpl",
		}

		if includeAMI {
			files["ami.pkr.hcl"] = "templates/ami.pkr.hcl.tmpl"
		}

		for filename, templatePath := range files {
			filePath := filepath.Join(templateDir, filename)
			if err := renderTemplate(templatePath, filePath, data); err != nil {
				logger.Errorf("Failed to create %s: %v", filename, err)
				return mcp.NewToolResultError(
					fmt.Sprintf("Failed to create %s: %v", filename, err)), nil
			}
		}

		logger.Infof("Successfully created template: %s", templateName)
		result := fmt.Sprintf(`Successfully created template '%s' at %s

Files created:
- plugins.pkr.hcl (Packer plugin requirements)
- locals.pkr.hcl (Local variables)
- variables.pkr.hcl (Template variables)
- docker.pkr.hcl (Docker build configuration)
- README.md (Template documentation)
%s

Next steps:
1. Initialize the template: task template-init TEMPLATE_NAME=%s
2. Customize the provisioning steps in docker.pkr.hcl
3. Validate the template: task template-validate TEMPLATE_NAME=%s
4. Build the template: task template-build TEMPLATE_NAME=%s
`, templateName, templateDir,
	func() string {
		if includeAMI {
			return "- ami.pkr.hcl (AWS AMI configuration)\n"
		}
		return ""
	}(), templateName, templateName, templateName)

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

