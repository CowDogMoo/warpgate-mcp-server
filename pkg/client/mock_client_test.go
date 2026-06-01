// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"errors"
	"testing"
)

func TestNewMockWarpgateClient(t *testing.T) {
	mock := NewMockWarpgateClient()

	if !mock.IsCLIAvailable() {
		t.Error("NewMockWarpgateClient should return client with CLI available")
	}

	if mock.GetCLIVersion() != "3.0.0" {
		t.Errorf("GetCLIVersion() = %q, want %q", mock.GetCLIVersion(), "3.0.0")
	}

	if mock.GetBinaryPath() != "/usr/local/bin/warpgate" {
		t.Errorf("GetBinaryPath() = %q, want %q", mock.GetBinaryPath(), "/usr/local/bin/warpgate")
	}

	if mock.GetRepoPath() != "/mock/warpgate" {
		t.Errorf("GetRepoPath() = %q, want %q", mock.GetRepoPath(), "/mock/warpgate")
	}
}

func TestMockWarpgateClientInterface(_ *testing.T) {
	// Ensure MockWarpgateClient implements WarpgateClientInterface
	var _ WarpgateClientInterface = (*MockWarpgateClient)(nil)
}

func TestMockWarpgateClient_ExecuteCLI(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ExecuteCLIResponse = "version: v3.0.0"

	result, err := mock.ExecuteCLI(context.Background(), "version")
	if err != nil {
		t.Errorf("ExecuteCLI returned unexpected error: %v", err)
	}

	if result != "version: v3.0.0" {
		t.Errorf("ExecuteCLI() = %q, want %q", result, "version: v3.0.0")
	}

	if len(mock.LastExecuteCLIArgs) != 1 || mock.LastExecuteCLIArgs[0] != "version" {
		t.Errorf("LastExecuteCLIArgs = %v, want [version]", mock.LastExecuteCLIArgs)
	}
}

func TestMockWarpgateClient_ExecuteCLIError(t *testing.T) {
	mock := NewMockWarpgateClient()
	expectedErr := errors.New("command failed")
	mock.ExecuteCLIError = expectedErr

	_, err := mock.ExecuteCLI(context.Background(), "build")
	if err != expectedErr {
		t.Errorf("ExecuteCLI error = %v, want %v", err, expectedErr)
	}
}

func TestMockWarpgateClient_WarpgateBuild(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.BuildResponse = "Build completed successfully"

	opts := BuildOptions{
		Target:        "container",
		Architectures: []string{"amd64"},
		Push:          true,
	}

	result, err := mock.WarpgateBuild(context.Background(), "attack-box", opts)
	if err != nil {
		t.Errorf("WarpgateBuild returned unexpected error: %v", err)
	}

	if result != "Build completed successfully" {
		t.Errorf("WarpgateBuild() = %q, want %q", result, "Build completed successfully")
	}

	if mock.LastBuildTemplate != "attack-box" {
		t.Errorf("LastBuildTemplate = %q, want %q", mock.LastBuildTemplate, "attack-box")
	}

	if mock.LastBuildOptions.Target != "container" {
		t.Errorf("LastBuildOptions.Target = %q, want %q", mock.LastBuildOptions.Target, "container")
	}
}

func TestMockWarpgateClient_WarpgateValidate(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ValidateResponse = "Validation passed"

	result, err := mock.WarpgateValidate(context.Background(), "/path/to/config.yaml", true)
	if err != nil {
		t.Errorf("WarpgateValidate returned unexpected error: %v", err)
	}

	if result != "Validation passed" {
		t.Errorf("WarpgateValidate() = %q, want %q", result, "Validation passed")
	}

	if mock.LastValidatePath != "/path/to/config.yaml" {
		t.Errorf("LastValidatePath = %q, want %q", mock.LastValidatePath, "/path/to/config.yaml")
	}
}

func TestMockWarpgateClient_WarpgateInit(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.InitResponse = "Template initialized"

	opts := InitOptions{
		OutputDir:    "/tmp/new-template",
		FromTemplate: "base-template",
	}

	result, err := mock.WarpgateInit(context.Background(), "my-template", opts)
	if err != nil {
		t.Errorf("WarpgateInit returned unexpected error: %v", err)
	}

	if result != "Template initialized" {
		t.Errorf("WarpgateInit() = %q, want %q", result, "Template initialized")
	}

	if mock.LastInitName != "my-template" {
		t.Errorf("LastInitName = %q, want %q", mock.LastInitName, "my-template")
	}

	if mock.LastInitOptions.OutputDir != "/tmp/new-template" {
		t.Errorf("LastInitOptions.OutputDir = %q, want %q", mock.LastInitOptions.OutputDir, "/tmp/new-template")
	}
}

func TestMockWarpgateClient_WarpgateTemplatesList(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.TemplatesListResponse = "attack-box\nsliver\natomic-red-team"

	result, err := mock.WarpgateTemplatesList(context.Background(), "local", "table")
	if err != nil {
		t.Errorf("WarpgateTemplatesList returned unexpected error: %v", err)
	}

	if result != "attack-box\nsliver\natomic-red-team" {
		t.Errorf("WarpgateTemplatesList() = %q, want templates list", result)
	}

	if mock.LastTemplatesListSource != "local" {
		t.Errorf("LastTemplatesListSource = %q, want %q", mock.LastTemplatesListSource, "local")
	}

	if mock.LastTemplatesListFormat != "table" {
		t.Errorf("LastTemplatesListFormat = %q, want %q", mock.LastTemplatesListFormat, "table")
	}
}

func TestMockWarpgateClient_WarpgateTemplatesInfo(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.TemplatesInfoResponse = "Template: attack-box\nDescription: Offensive security tools"

	result, err := mock.WarpgateTemplatesInfo(context.Background(), "attack-box")
	if err != nil {
		t.Errorf("WarpgateTemplatesInfo returned unexpected error: %v", err)
	}

	if mock.LastTemplatesInfoTemplate != "attack-box" {
		t.Errorf("LastTemplatesInfoTemplate = %q, want %q", mock.LastTemplatesInfoTemplate, "attack-box")
	}

	if result == "" {
		t.Error("WarpgateTemplatesInfo should return template info")
	}
}

func TestMockWarpgateClient_WarpgateTemplatesAdd(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.TemplatesAddResponse = "Template source added"

	result, err := mock.WarpgateTemplatesAdd(context.Background(), "https://github.com/example/templates", "my-templates")
	if err != nil {
		t.Errorf("WarpgateTemplatesAdd returned unexpected error: %v", err)
	}

	if result != "Template source added" {
		t.Errorf("WarpgateTemplatesAdd() = %q, want %q", result, "Template source added")
	}

	if mock.LastTemplatesAddSource != "https://github.com/example/templates" {
		t.Errorf("LastTemplatesAddSource = %q, want %q", mock.LastTemplatesAddSource, "https://github.com/example/templates")
	}

	if mock.LastTemplatesAddName != "my-templates" {
		t.Errorf("LastTemplatesAddName = %q, want %q", mock.LastTemplatesAddName, "my-templates")
	}
}

func TestMockWarpgateClient_WarpgateTemplatesRemove(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.TemplatesRemoveResponse = "Template source removed"

	result, err := mock.WarpgateTemplatesRemove(context.Background(), "my-templates")
	if err != nil {
		t.Errorf("WarpgateTemplatesRemove returned unexpected error: %v", err)
	}

	if result != "Template source removed" {
		t.Errorf("WarpgateTemplatesRemove() = %q, want %q", result, "Template source removed")
	}

	if mock.LastTemplatesRemoveName != "my-templates" {
		t.Errorf("LastTemplatesRemoveName = %q, want %q", mock.LastTemplatesRemoveName, "my-templates")
	}
}

func TestMockWarpgateClient_WarpgateManifestsCreate(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ManifestsCreateResponse = "Manifest created"

	opts := ManifestsCreateOptions{
		Name:     "my-manifest",
		Registry: "ghcr.io/cowdogmoo",
		Tags:     []string{"latest", "v1.0.0"},
		DryRun:   true,
	}
	result, err := mock.WarpgateManifestsCreate(context.Background(), opts)
	if err != nil {
		t.Errorf("WarpgateManifestsCreate returned unexpected error: %v", err)
	}

	if result != "Manifest created" {
		t.Errorf("WarpgateManifestsCreate() = %q, want %q", result, "Manifest created")
	}

	if mock.LastManifestsCreateOpts.Name != "my-manifest" {
		t.Errorf("LastManifestsCreateOpts.Name = %q, want %q", mock.LastManifestsCreateOpts.Name, "my-manifest")
	}

	if mock.LastManifestsCreateOpts.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("LastManifestsCreateOpts.Registry = %q, want %q", mock.LastManifestsCreateOpts.Registry, "ghcr.io/cowdogmoo")
	}

	if len(mock.LastManifestsCreateOpts.Tags) != 2 {
		t.Errorf("LastManifestsCreateOpts.Tags length = %d, want 2", len(mock.LastManifestsCreateOpts.Tags))
	}

	if !mock.LastManifestsCreateOpts.DryRun {
		t.Error("LastManifestsCreateOpts.DryRun should be true")
	}
}

func TestMockWarpgateClient_WarpgateManifestsList(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ManifestsListResponse = "latest\nv1.0.0\nv1.0.1"

	opts := ManifestsListOptions{
		Name:     "attack-box",
		Registry: "ghcr.io/cowdogmoo",
	}

	result, err := mock.WarpgateManifestsList(context.Background(), opts)
	if err != nil {
		t.Errorf("WarpgateManifestsList returned unexpected error: %v", err)
	}

	if result != "latest\nv1.0.0\nv1.0.1" {
		t.Errorf("WarpgateManifestsList() = %q, want tags list", result)
	}

	if mock.LastManifestsListOptions.Name != "attack-box" {
		t.Errorf("LastManifestsListOptions.Name = %q, want %q", mock.LastManifestsListOptions.Name, "attack-box")
	}

	if mock.LastManifestsListOptions.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("LastManifestsListOptions.Registry = %q, want %q", mock.LastManifestsListOptions.Registry, "ghcr.io/cowdogmoo")
	}
}

func TestMockWarpgateClient_WarpgateManifestsInspect(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ManifestsInspectResponse = "Manifest: attack-box:latest\nArchitectures: amd64, arm64"

	opts := ManifestsInspectOptions{
		Name:     "attack-box",
		Registry: "ghcr.io/cowdogmoo",
		Tags:     []string{"latest"},
	}

	result, err := mock.WarpgateManifestsInspect(context.Background(), opts)
	if err != nil {
		t.Errorf("WarpgateManifestsInspect returned unexpected error: %v", err)
	}

	if result == "" {
		t.Error("WarpgateManifestsInspect should return manifest info")
	}

	if mock.LastManifestsInspectOpts.Name != "attack-box" {
		t.Errorf("LastManifestsInspectOpts.Name = %q, want %q", mock.LastManifestsInspectOpts.Name, "attack-box")
	}
}

func TestMockWarpgateClient_WarpgateConfigGet(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ConfigGetResponse = "ghcr.io/cowdogmoo"

	result, err := mock.WarpgateConfigGet(context.Background(), "registry")
	if err != nil {
		t.Errorf("WarpgateConfigGet returned unexpected error: %v", err)
	}

	if result != "ghcr.io/cowdogmoo" {
		t.Errorf("WarpgateConfigGet() = %q, want %q", result, "ghcr.io/cowdogmoo")
	}

	if mock.LastConfigGetKey != "registry" {
		t.Errorf("LastConfigGetKey = %q, want %q", mock.LastConfigGetKey, "registry")
	}
}

func TestMockWarpgateClient_WarpgateConfigSet(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ConfigSetResponse = "Configuration updated"

	result, err := mock.WarpgateConfigSet(context.Background(), "registry", "ghcr.io/test")
	if err != nil {
		t.Errorf("WarpgateConfigSet returned unexpected error: %v", err)
	}

	if result != "Configuration updated" {
		t.Errorf("WarpgateConfigSet() = %q, want %q", result, "Configuration updated")
	}

	if mock.LastConfigSetKey != "registry" {
		t.Errorf("LastConfigSetKey = %q, want %q", mock.LastConfigSetKey, "registry")
	}

	if mock.LastConfigSetValue != "ghcr.io/test" {
		t.Errorf("LastConfigSetValue = %q, want %q", mock.LastConfigSetValue, "ghcr.io/test")
	}
}

func TestMockWarpgateClient_WarpgateConfigShow(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ConfigShowResponse = "registry: ghcr.io/cowdogmoo\nregion: us-east-1"

	result, err := mock.WarpgateConfigShow(context.Background())
	if err != nil {
		t.Errorf("WarpgateConfigShow returned unexpected error: %v", err)
	}

	if result != "registry: ghcr.io/cowdogmoo\nregion: us-east-1" {
		t.Errorf("WarpgateConfigShow() = %q, want config output", result)
	}
}

func TestMockWarpgateClient_WarpgateConvert(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ConvertResponse = "Conversion completed"

	result, err := mock.WarpgateConvert(context.Background(), "/path/to/packer.json", "/path/to/output")
	if err != nil {
		t.Errorf("WarpgateConvert returned unexpected error: %v", err)
	}

	if result != "Conversion completed" {
		t.Errorf("WarpgateConvert() = %q, want %q", result, "Conversion completed")
	}

	if mock.LastConvertSource != "/path/to/packer.json" {
		t.Errorf("LastConvertSource = %q, want %q", mock.LastConvertSource, "/path/to/packer.json")
	}

	if mock.LastConvertOutput != "/path/to/output" {
		t.Errorf("LastConvertOutput = %q, want %q", mock.LastConvertOutput, "/path/to/output")
	}
}

func TestMockWarpgateClient_WarpgateValidateConfig(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ValidateConfigResponse = "Configuration is valid"

	result, err := mock.WarpgateValidateConfig(context.Background(), "/path/to/warpgate.yaml")
	if err != nil {
		t.Errorf("WarpgateValidateConfig returned unexpected error: %v", err)
	}

	if result != "Configuration is valid" {
		t.Errorf("WarpgateValidateConfig() = %q, want %q", result, "Configuration is valid")
	}

	if mock.LastValidateConfigPath != "/path/to/warpgate.yaml" {
		t.Errorf("LastValidateConfigPath = %q, want %q", mock.LastValidateConfigPath, "/path/to/warpgate.yaml")
	}
}

func TestMockWarpgateClient_CLINotAvailable(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.CLIAvailable = false

	if mock.IsCLIAvailable() {
		t.Error("IsCLIAvailable() should return false when CLIAvailable is false")
	}
}

func TestMockWarpgateClient_ErrorResponses(t *testing.T) {
	mock := NewMockWarpgateClient()
	expectedErr := errors.New("test error")

	tests := []struct {
		name    string
		setup   func()
		execute func() error
	}{
		{
			name: "BuildError",
			setup: func() {
				mock.BuildError = expectedErr
			},
			execute: func() error {
				_, err := mock.WarpgateBuild(context.Background(), "test", BuildOptions{})
				return err
			},
		},
		{
			name: "ValidateError",
			setup: func() {
				mock.ValidateError = expectedErr
			},
			execute: func() error {
				_, err := mock.WarpgateValidate(context.Background(), "path", false)
				return err
			},
		},
		{
			name: "InitError",
			setup: func() {
				mock.InitError = expectedErr
			},
			execute: func() error {
				_, err := mock.WarpgateInit(context.Background(), "name", InitOptions{})
				return err
			},
		},
		{
			name: "TemplatesListError",
			setup: func() {
				mock.TemplatesListError = expectedErr
			},
			execute: func() error {
				_, err := mock.WarpgateTemplatesList(context.Background(), "", "")
				return err
			},
		},
		{
			name: "ConfigShowError",
			setup: func() {
				mock.ConfigShowError = expectedErr
			},
			execute: func() error {
				_, err := mock.WarpgateConfigShow(context.Background())
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mock = NewMockWarpgateClient()
			tt.setup()

			err := tt.execute()
			if err != expectedErr {
				t.Errorf("%s: error = %v, want %v", tt.name, err, expectedErr)
			}
		})
	}
}
