// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsValidTemplateName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid lowercase", "mytemplate", true},
		{"valid with hyphen", "my-template", true},
		{"valid with numbers", "template123", true},
		{"valid mixed", "my-template-123", true},
		{"invalid uppercase", "MyTemplate", false},
		{"invalid underscore", "my_template", false},
		{"invalid space", "my template", false},
		{"invalid special chars", "my@template", false},
		{"empty string", "", false},
		{"single char", "a", true},
		{"single number", "1", true},
		{"starts with hyphen", "-template", true}, // allowed by current impl
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTemplateName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidTemplateName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEmbeddedTemplatesExist(t *testing.T) {
	// Test that embedded templates can be read
	templates := []string{
		"templates/warpgate.yaml.tmpl",
		"templates/warpgate-README.md.tmpl",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := templateFS.ReadFile(tmpl)
			if err != nil {
				t.Errorf("Failed to read embedded template %s: %v", tmpl, err)
				return
			}
			if len(content) == 0 {
				t.Errorf("Embedded template %s is empty", tmpl)
			}
		})
	}
}

func TestEmbeddedWarpgateYamlTemplate(t *testing.T) {
	content, err := templateFS.ReadFile("templates/warpgate.yaml.tmpl")
	if err != nil {
		t.Fatalf("Failed to read warpgate.yaml.tmpl: %v", err)
	}

	contentStr := string(content)

	// Verify key template markers exist
	expectedMarkers := []string{
		"{{.TemplateName}}",
		"{{.Description}}",
		"{{.BaseImage}}",
		"{{.BaseImageVersion}}",
		"metadata:",
		"provisioners:",
		"targets:",
	}

	for _, marker := range expectedMarkers {
		if !strings.Contains(contentStr, marker) {
			t.Errorf("warpgate.yaml.tmpl missing expected content: %s", marker)
		}
	}
}

func TestEmbeddedReadmeTemplate(t *testing.T) {
	content, err := templateFS.ReadFile("templates/warpgate-README.md.tmpl")
	if err != nil {
		t.Fatalf("Failed to read warpgate-README.md.tmpl: %v", err)
	}

	contentStr := string(content)

	// Verify key template markers exist
	expectedMarkers := []string{
		"{{.Title}}",
		"{{.Description}}",
		"warpgate validate",
		"warpgate build",
	}

	for _, marker := range expectedMarkers {
		if !strings.Contains(contentStr, marker) {
			t.Errorf("warpgate-README.md.tmpl missing expected content: %s", marker)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "test-template",
		Description:      "A test template",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		IncludeAMI:       false,
		Title:            "Test-Template",
		Platforms:        []string{"linux/amd64"},
	}

	// Test rendering warpgate.yaml
	outputPath := filepath.Join(tmpDir, "warpgate.yaml")
	err = renderTemplate("templates/warpgate.yaml.tmpl", outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile(outputPath) //nolint:gosec // G304: test file reading
	if err != nil {
		t.Fatalf("Failed to read rendered template: %v", err)
	}

	contentStr := string(content)

	// Verify template variables were substituted
	if !strings.Contains(contentStr, "test-template") {
		t.Error("Rendered template missing template name")
	}
	if !strings.Contains(contentStr, "A test template") {
		t.Error("Rendered template missing description")
	}
	if !strings.Contains(contentStr, "ubuntu:22.04") {
		t.Error("Rendered template missing base image")
	}
	if !strings.Contains(contentStr, "linux/amd64") {
		t.Error("Rendered template missing platform")
	}

	// Verify no unsubstituted template markers remain (except for escaped ones)
	if strings.Contains(contentStr, "{{.TemplateName}}") {
		t.Error("Rendered template contains unsubstituted {{.TemplateName}}")
	}
}

func TestRenderTemplateWithAMI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "warpgate-test-ami-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "ami-test",
		Description:      "AMI test template",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		IncludeAMI:       true,
		Title:            "AMI-Test",
		Platforms:        nil, // use defaults
	}

	outputPath := filepath.Join(tmpDir, "warpgate.yaml")
	err = renderTemplate("templates/warpgate.yaml.tmpl", outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	content, err := os.ReadFile(outputPath) //nolint:gosec // G304: test file reading
	if err != nil {
		t.Fatalf("Failed to read rendered template: %v", err)
	}

	contentStr := string(content)

	// Verify AMI configuration is included
	if !strings.Contains(contentStr, "type: ami") {
		t.Error("Rendered template with IncludeAMI=true should contain AMI target")
	}
	if !strings.Contains(contentStr, "region:") {
		t.Error("Rendered template with IncludeAMI=true should contain region")
	}
}

func TestTemplateDataStruct(t *testing.T) {
	data := templateData{
		TemplateName:     "test",
		Description:      "desc",
		BaseImage:        "alpine",
		BaseImageVersion: "3.18",
		IncludeAMI:       true,
		Title:            "Test",
		Platforms:        []string{"linux/amd64", "linux/arm64"},
	}

	if data.TemplateName != "test" {
		t.Errorf("templateData.TemplateName = %q, want %q", data.TemplateName, "test")
	}
	if len(data.Platforms) != 2 {
		t.Errorf("templateData.Platforms length = %d, want 2", len(data.Platforms))
	}
}

func TestRenderTemplateErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "test",
		Description:      "Test template",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		IncludeAMI:       false,
		Title:            "Test",
		Platforms:        []string{"linux/amd64"},
	}

	// Test with invalid template path
	err = renderTemplate("nonexistent/template.tmpl", filepath.Join(tmpDir, "output.yaml"), data)
	if err == nil {
		t.Error("renderTemplate should fail with invalid template path")
	}

	// Test with invalid output path (directory doesn't exist)
	err = renderTemplate("templates/warpgate.yaml.tmpl", "/nonexistent/dir/output.yaml", data)
	if err == nil {
		t.Error("renderTemplate should fail with invalid output path")
	}
}

func TestRenderTemplateReadmeWithAMI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "warpgate-test-ami-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "test-ami",
		Description:      "Template with AMI support",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		IncludeAMI:       true,
		Title:            "Test-AMI",
		Platforms:        []string{"linux/amd64", "linux/arm64"},
	}

	// Test rendering README with AMI
	outputPath := filepath.Join(tmpDir, "README.md")
	err = renderTemplate("templates/warpgate-README.md.tmpl", outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render README template: %v", err)
	}

	content, err := os.ReadFile(outputPath) //nolint:gosec // G304: test file reading
	if err != nil {
		t.Fatalf("Failed to read rendered README: %v", err)
	}

	contentStr := string(content)

	// Verify template variables were substituted
	if !strings.Contains(contentStr, "Test-AMI") {
		t.Error("Rendered README missing title")
	}
	if !strings.Contains(contentStr, "Template with AMI support") {
		t.Error("Rendered README missing description")
	}
}

func TestRenderTemplateDefaultPlatforms(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "warpgate-test-default-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "default-platforms",
		Description:      "Test default platforms",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		IncludeAMI:       false,
		Title:            "Default-Platforms",
		Platforms:        nil, // Empty platforms
	}

	outputPath := filepath.Join(tmpDir, "warpgate.yaml")
	err = renderTemplate("templates/warpgate.yaml.tmpl", outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// File should be created even with nil platforms
	_, err = os.Stat(outputPath)
	if err != nil {
		t.Errorf("Output file should exist: %v", err)
	}
}

func TestIsValidTemplateNameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"long name", "this-is-a-very-long-template-name-that-should-still-be-valid-123", true},
		{"only numbers", "123456", true},
		{"only hyphens", "---", true},
		{"mixed valid chars", "a1-b2-c3", true},
		{"unicode", "テスト", false},
		{"emoji", "test🚀", false},
		{"newline", "test\nname", false},
		{"tab", "test\tname", false},
		{"path separator", "test/name", false},
		{"backslash", "test\\name", false},
		{"colon", "test:name", false},
		{"period", "test.name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTemplateName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidTemplateName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEmbeddedTemplatesContent(t *testing.T) {
	// Test that embedded templates have expected structure
	templates := []struct {
		path     string
		required []string
	}{
		{
			path:     "templates/warpgate.yaml.tmpl",
			required: []string{"metadata:", "provisioners:", "targets:"},
		},
		{
			path:     "templates/warpgate-README.md.tmpl",
			required: []string{"#", "warpgate"},
		},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.path, func(t *testing.T) {
			content, err := templateFS.ReadFile(tmpl.path)
			if err != nil {
				t.Fatalf("Failed to read template %s: %v", tmpl.path, err)
			}

			contentStr := string(content)
			for _, req := range tmpl.required {
				if !strings.Contains(contentStr, req) {
					t.Errorf("Template %s missing required content: %s", tmpl.path, req)
				}
			}
		})
	}
}

func TestTemplateDataAllFields(t *testing.T) {
	// Test all combinations of templateData fields
	testCases := []struct {
		name string
		data templateData
	}{
		{
			name: "minimal",
			data: templateData{
				TemplateName: "min",
				Description:  "minimal",
			},
		},
		{
			name: "full",
			data: templateData{
				TemplateName:     "full",
				Description:      "full template",
				BaseImage:        "alpine",
				BaseImageVersion: "3.19",
				IncludeAMI:       true,
				Title:            "Full",
				Platforms:        []string{"linux/amd64", "linux/arm64", "linux/arm/v7"},
			},
		},
		{
			name: "empty platforms",
			data: templateData{
				TemplateName:     "empty-plat",
				Description:      "empty platforms",
				BaseImage:        "ubuntu",
				BaseImageVersion: "24.04",
				IncludeAMI:       false,
				Title:            "Empty-Plat",
				Platforms:        []string{},
			},
		},
	}

	tmpDir, err := os.MkdirTemp("", "warpgate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, tc.name+".yaml")
			err := renderTemplate("templates/warpgate.yaml.tmpl", outputPath, tc.data)
			if err != nil {
				t.Errorf("renderTemplate failed for %s: %v", tc.name, err)
			}
		})
	}
}

func TestRenderTemplateFilePermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "warpgate-test-perms-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data := templateData{
		TemplateName:     "perms-test",
		Description:      "Test file permissions",
		BaseImage:        "ubuntu",
		BaseImageVersion: "22.04",
		Title:            "Perms-Test",
	}

	outputPath := filepath.Join(tmpDir, "warpgate.yaml")
	err = renderTemplate("templates/warpgate.yaml.tmpl", outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Check that file was created and is readable
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file should not be empty")
	}
}
