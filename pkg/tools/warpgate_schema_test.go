// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"testing"
)

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"simple version", "1.2.3", true},
		{"with v prefix", "v1.2.3", true},
		{"two parts", "1.2", true},
		{"four parts", "1.2.3.4", false},
		{"one part", "1", false},
		{"invalid chars", "abc", false},
		{"empty", "", false},
		{"with spaces", "1.2 .3", false},
		{"with letters", "1.2.a", false},
		{"zero version", "0.0.0", true},
		{"large numbers", "10.20.30", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVersion(tt.version)
			if result != tt.expected {
				t.Errorf("isValidVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestWarpgateTemplateConfigStruct(t *testing.T) {
	// Test that the config struct fields are properly defined
	config := WarpgateTemplateConfig{
		Name:        "test-template",
		Description: "A test template",
		Version:     "v1.0.0",
		Maintainers: []string{"maintainer@example.com"},
		BaseImage: &BaseImageConfig{
			Name: "ubuntu",
			Tag:  "22.04",
		},
		Targets: &TargetsConfig{
			Container: &ContainerTargetConfig{
				Enabled:       true,
				Architectures: []string{"amd64", "arm64"},
			},
			AMI: &AMITargetConfig{
				Enabled:      true,
				Region:       "us-east-1",
				InstanceType: "t3.medium",
			},
		},
		Provisioners: []ProvisionerConfig{
			{
				Type:   "shell",
				Name:   "update-packages",
				Inline: []string{"apt-get update", "apt-get upgrade -y"},
			},
			{
				Type:        "file",
				Name:        "copy-config",
				Source:      "config.yaml",
				Destination: "/etc/app/config.yaml",
			},
		},
		Variables: map[string]interface{}{
			"key": "value",
		},
	}

	if config.Name != "test-template" {
		t.Errorf("config.Name = %q, want %q", config.Name, "test-template")
	}

	if config.BaseImage.Name != "ubuntu" {
		t.Errorf("config.BaseImage.Name = %q, want %q", config.BaseImage.Name, "ubuntu")
	}

	if !config.Targets.Container.Enabled {
		t.Error("config.Targets.Container.Enabled should be true")
	}

	if len(config.Provisioners) != 2 {
		t.Errorf("len(config.Provisioners) = %d, want 2", len(config.Provisioners))
	}

	if config.Provisioners[0].Type != "shell" {
		t.Errorf("config.Provisioners[0].Type = %q, want %q", config.Provisioners[0].Type, "shell")
	}
}

func TestBaseImageConfig(t *testing.T) {
	config := BaseImageConfig{
		Name: "alpine",
		Tag:  "3.18",
	}

	if config.Name != "alpine" {
		t.Errorf("BaseImageConfig.Name = %q, want %q", config.Name, "alpine")
	}

	if config.Tag != "3.18" {
		t.Errorf("BaseImageConfig.Tag = %q, want %q", config.Tag, "3.18")
	}
}

func TestContainerTargetConfig(t *testing.T) {
	config := ContainerTargetConfig{
		Enabled:       true,
		Architectures: []string{"amd64", "arm64"},
	}

	if !config.Enabled {
		t.Error("ContainerTargetConfig.Enabled should be true")
	}

	if len(config.Architectures) != 2 {
		t.Errorf("len(config.Architectures) = %d, want 2", len(config.Architectures))
	}

	if config.Architectures[0] != "amd64" {
		t.Errorf("config.Architectures[0] = %q, want %q", config.Architectures[0], "amd64")
	}
}

func TestAMITargetConfig(t *testing.T) {
	config := AMITargetConfig{
		Enabled:      true,
		Region:       "us-west-2",
		InstanceType: "t3.large",
	}

	if !config.Enabled {
		t.Error("AMITargetConfig.Enabled should be true")
	}

	if config.Region != "us-west-2" {
		t.Errorf("AMITargetConfig.Region = %q, want %q", config.Region, "us-west-2")
	}

	if config.InstanceType != "t3.large" {
		t.Errorf("AMITargetConfig.InstanceType = %q, want %q", config.InstanceType, "t3.large")
	}
}

func TestProvisionerConfig(t *testing.T) {
	tests := []struct {
		name   string
		config ProvisionerConfig
	}{
		{
			name: "shell provisioner with inline",
			config: ProvisionerConfig{
				Type:   "shell",
				Name:   "install-packages",
				Inline: []string{"apt-get update", "apt-get install -y curl"},
			},
		},
		{
			name: "shell provisioner with script",
			config: ProvisionerConfig{
				Type:   "shell",
				Name:   "run-script",
				Script: "/scripts/setup.sh",
			},
		},
		{
			name: "file provisioner",
			config: ProvisionerConfig{
				Type:        "file",
				Name:        "copy-file",
				Source:      "config.json",
				Destination: "/etc/app/config.json",
			},
		},
		{
			name: "ansible provisioner",
			config: ProvisionerConfig{
				Type: "ansible",
				Name: "run-playbook",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Type == "" {
				t.Error("ProvisionerConfig.Type should not be empty")
			}
		})
	}
}

func TestTargetsConfig(t *testing.T) {
	config := TargetsConfig{
		Container: &ContainerTargetConfig{
			Enabled:       true,
			Architectures: []string{"amd64"},
		},
		AMI: &AMITargetConfig{
			Enabled:      false,
			Region:       "us-east-1",
			InstanceType: "t3.medium",
		},
	}

	if config.Container == nil {
		t.Error("TargetsConfig.Container should not be nil")
	}

	if config.AMI == nil {
		t.Error("TargetsConfig.AMI should not be nil")
	}

	if !config.Container.Enabled {
		t.Error("Container target should be enabled")
	}

	if config.AMI.Enabled {
		t.Error("AMI target should be disabled")
	}
}

func TestNilTargetsConfig(t *testing.T) {
	// Test that nil targets don't cause issues
	config := WarpgateTemplateConfig{
		Name:    "minimal-template",
		Targets: nil,
	}

	if config.Targets != nil {
		t.Error("config.Targets should be nil")
	}
}

func TestEmptyProvisionersList(t *testing.T) {
	config := WarpgateTemplateConfig{
		Name:         "minimal-template",
		Provisioners: []ProvisionerConfig{},
	}

	if len(config.Provisioners) != 0 {
		t.Errorf("len(config.Provisioners) = %d, want 0", len(config.Provisioners))
	}
}

func TestVariablesMap(t *testing.T) {
	config := WarpgateTemplateConfig{
		Name: "test-template",
		Variables: map[string]interface{}{
			"string_var": "value",
			"int_var":    42,
			"bool_var":   true,
			"float_var":  3.14,
			"nested_var": map[string]interface{}{"key": "value"},
		},
	}

	if config.Variables["string_var"] != "value" {
		t.Errorf("Variables[string_var] = %v, want %q", config.Variables["string_var"], "value")
	}

	if config.Variables["int_var"] != 42 {
		t.Errorf("Variables[int_var] = %v, want 42", config.Variables["int_var"])
	}

	if config.Variables["bool_var"] != true {
		t.Errorf("Variables[bool_var] = %v, want true", config.Variables["bool_var"])
	}
}
