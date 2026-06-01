// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

// Package client provides the warpgate CLI client interface and implementations.
package client

import "context"

// WarpgateClientInterface defines the interface for warpgate client operations
// This allows for easy mocking in tests
type WarpgateClientInterface interface {
	// CLI detection
	IsCLIAvailable() bool
	GetCLIVersion() string
	GetBinaryPath() string
	GetRepoPath() string

	// CLI execution
	ExecuteCLI(ctx context.Context, args ...string) (string, error)
	ExecuteCLIWithWorkdir(ctx context.Context, workdir string, args ...string) (string, error)
	ExecuteCLIStreaming(ctx context.Context, callback OutputCallback, args ...string) (string, error)

	// Build operations
	WarpgateBuild(ctx context.Context, template string, opts BuildOptions) (string, error)
	WarpgateBuildStreaming(ctx context.Context, template string, opts BuildOptions, callback OutputCallback) (string, error)
	WarpgateValidate(ctx context.Context, configPath string, syntaxOnly bool) (string, error)
	WarpgateInit(ctx context.Context, name string, opts InitOptions) (string, error)

	// Template operations
	WarpgateTemplatesList(ctx context.Context, source, format string) (string, error)
	WarpgateTemplatesInfo(ctx context.Context, template string) (string, error)
	WarpgateTemplatesAdd(ctx context.Context, source string, name string) (string, error)
	WarpgateTemplatesRemove(ctx context.Context, nameOrPath string) (string, error)

	// Manifest operations
	WarpgateManifestsCreate(ctx context.Context, opts ManifestsCreateOptions) (string, error)
	WarpgateManifestsList(ctx context.Context, opts ManifestsListOptions) (string, error)
	WarpgateManifestsInspect(ctx context.Context, opts ManifestsInspectOptions) (string, error)

	// Config operations
	WarpgateConfigGet(ctx context.Context, key string) (string, error)
	WarpgateConfigSet(ctx context.Context, key, value string) (string, error)
	WarpgateConfigShow(ctx context.Context) (string, error)

	// Conversion
	WarpgateConvert(ctx context.Context, source, output string) (string, error)
	WarpgateValidateConfig(ctx context.Context, configPath string) (string, error)

	// Registry operations
	RegistryDelete(ctx context.Context, opts RegistryDeleteOptions) (string, error)
	RegistryCopy(ctx context.Context, opts RegistryCopyOptions) (string, error)
}

// Ensure WarpgateClient implements the interface
var _ WarpgateClientInterface = (*WarpgateClient)(nil)

// MockWarpgateClient is a mock implementation for testing
type MockWarpgateClient struct {
	// Configuration
	CLIAvailable bool
	CLIVersion   string
	BinaryPath   string
	RepoPath     string

	// Mock responses - can be set per test
	ExecuteCLIResponse       string
	ExecuteCLIError          error
	ExecuteCLIStreamingLines []string
	BuildResponse            string
	BuildError               error
	BuildStreamingLines      []string
	ValidateResponse         string
	ValidateError            error
	InitResponse             string
	InitError                error
	TemplatesListResponse    string
	TemplatesListError       error
	TemplatesInfoResponse    string
	TemplatesInfoError       error
	TemplatesAddResponse     string
	TemplatesAddError        error
	TemplatesRemoveResponse  string
	TemplatesRemoveError     error
	ManifestsCreateResponse  string
	ManifestsCreateError     error
	ManifestsListResponse    string
	ManifestsListError       error
	ManifestsInspectResponse string
	ManifestsInspectError    error
	ConfigGetResponse        string
	ConfigGetError           error
	ConfigSetResponse        string
	ConfigSetError           error
	ConfigShowResponse       string
	ConfigShowError          error
	ConvertResponse          string
	ConvertError             error
	ValidateConfigResponse   string
	ValidateConfigError      error
	RegistryDeleteResponse   string
	RegistryDeleteError      error
	RegistryCopyResponse     string
	RegistryCopyError        error

	// Call tracking
	LastExecuteCLIArgs        []string
	LastBuildTemplate         string
	LastBuildOptions          BuildOptions
	LastBuildCallback         OutputCallback
	LastValidatePath          string
	LastInitName              string
	LastInitOptions           InitOptions
	LastTemplatesListSource   string
	LastTemplatesListFormat   string
	LastTemplatesInfoTemplate string
	LastTemplatesAddSource    string
	LastTemplatesAddName      string
	LastTemplatesRemoveName   string
	LastManifestsCreateOpts   ManifestsCreateOptions
	LastManifestsListOptions  ManifestsListOptions
	LastManifestsInspectOpts  ManifestsInspectOptions
	LastConfigGetKey          string
	LastConfigSetKey          string
	LastConfigSetValue        string
	LastConvertSource         string
	LastConvertOutput         string
	LastValidateConfigPath    string
	LastRegistryDeleteOptions RegistryDeleteOptions
	LastRegistryCopyOptions   RegistryCopyOptions
}

// NewMockWarpgateClient creates a new mock client with default values
func NewMockWarpgateClient() *MockWarpgateClient {
	return &MockWarpgateClient{
		CLIAvailable: true,
		CLIVersion:   "3.0.0",
		BinaryPath:   "/usr/local/bin/warpgate",
		RepoPath:     "/mock/warpgate",
	}
}

// IsCLIAvailable returns the mock CLI availability
func (m *MockWarpgateClient) IsCLIAvailable() bool {
	return m.CLIAvailable
}

// GetCLIVersion returns the mock CLI version
func (m *MockWarpgateClient) GetCLIVersion() string {
	return m.CLIVersion
}

// GetBinaryPath returns the mock binary path
func (m *MockWarpgateClient) GetBinaryPath() string {
	return m.BinaryPath
}

// GetRepoPath returns the mock repo path
func (m *MockWarpgateClient) GetRepoPath() string {
	return m.RepoPath
}

// ExecuteCLI mocks CLI execution
func (m *MockWarpgateClient) ExecuteCLI(_ context.Context, args ...string) (string, error) {
	m.LastExecuteCLIArgs = args
	return m.ExecuteCLIResponse, m.ExecuteCLIError
}

// ExecuteCLIWithWorkdir mocks CLI execution with workdir
func (m *MockWarpgateClient) ExecuteCLIWithWorkdir(_ context.Context, _ string, args ...string) (string, error) {
	m.LastExecuteCLIArgs = args
	return m.ExecuteCLIResponse, m.ExecuteCLIError
}

// WarpgateBuild mocks the build command
func (m *MockWarpgateClient) WarpgateBuild(_ context.Context, template string, opts BuildOptions) (string, error) {
	m.LastBuildTemplate = template
	m.LastBuildOptions = opts
	return m.BuildResponse, m.BuildError
}

// WarpgateValidate mocks the validate command
func (m *MockWarpgateClient) WarpgateValidate(_ context.Context, configPath string, _ bool) (string, error) {
	m.LastValidatePath = configPath
	return m.ValidateResponse, m.ValidateError
}

// WarpgateInit mocks the init command
func (m *MockWarpgateClient) WarpgateInit(_ context.Context, name string, opts InitOptions) (string, error) {
	m.LastInitName = name
	m.LastInitOptions = opts
	return m.InitResponse, m.InitError
}

// WarpgateTemplatesList mocks the templates list command
func (m *MockWarpgateClient) WarpgateTemplatesList(_ context.Context, source, format string) (string, error) {
	m.LastTemplatesListSource = source
	m.LastTemplatesListFormat = format
	return m.TemplatesListResponse, m.TemplatesListError
}

// WarpgateTemplatesInfo mocks the templates info command
func (m *MockWarpgateClient) WarpgateTemplatesInfo(_ context.Context, template string) (string, error) {
	m.LastTemplatesInfoTemplate = template
	return m.TemplatesInfoResponse, m.TemplatesInfoError
}

// WarpgateTemplatesAdd mocks the templates add command
func (m *MockWarpgateClient) WarpgateTemplatesAdd(_ context.Context, source string, name string) (string, error) {
	m.LastTemplatesAddSource = source
	m.LastTemplatesAddName = name
	return m.TemplatesAddResponse, m.TemplatesAddError
}

// WarpgateTemplatesRemove mocks the templates remove command
func (m *MockWarpgateClient) WarpgateTemplatesRemove(_ context.Context, nameOrPath string) (string, error) {
	m.LastTemplatesRemoveName = nameOrPath
	return m.TemplatesRemoveResponse, m.TemplatesRemoveError
}

// WarpgateManifestsCreate mocks the manifests create command
func (m *MockWarpgateClient) WarpgateManifestsCreate(_ context.Context, opts ManifestsCreateOptions) (string, error) {
	m.LastManifestsCreateOpts = opts
	return m.ManifestsCreateResponse, m.ManifestsCreateError
}

// WarpgateManifestsList mocks the manifests list command
func (m *MockWarpgateClient) WarpgateManifestsList(_ context.Context, opts ManifestsListOptions) (string, error) {
	m.LastManifestsListOptions = opts
	return m.ManifestsListResponse, m.ManifestsListError
}

// WarpgateManifestsInspect mocks the manifests inspect command
func (m *MockWarpgateClient) WarpgateManifestsInspect(_ context.Context, opts ManifestsInspectOptions) (string, error) {
	m.LastManifestsInspectOpts = opts
	return m.ManifestsInspectResponse, m.ManifestsInspectError
}

// WarpgateConfigGet mocks the config get command
func (m *MockWarpgateClient) WarpgateConfigGet(_ context.Context, key string) (string, error) {
	m.LastConfigGetKey = key
	return m.ConfigGetResponse, m.ConfigGetError
}

// WarpgateConfigSet mocks the config set command
func (m *MockWarpgateClient) WarpgateConfigSet(_ context.Context, key, value string) (string, error) {
	m.LastConfigSetKey = key
	m.LastConfigSetValue = value
	return m.ConfigSetResponse, m.ConfigSetError
}

// WarpgateConfigShow mocks the config show command
func (m *MockWarpgateClient) WarpgateConfigShow(_ context.Context) (string, error) {
	return m.ConfigShowResponse, m.ConfigShowError
}

// WarpgateConvert mocks the convert command
func (m *MockWarpgateClient) WarpgateConvert(_ context.Context, source, output string) (string, error) {
	m.LastConvertSource = source
	m.LastConvertOutput = output
	return m.ConvertResponse, m.ConvertError
}

// WarpgateValidateConfig mocks the validate config command
func (m *MockWarpgateClient) WarpgateValidateConfig(_ context.Context, configPath string) (string, error) {
	m.LastValidateConfigPath = configPath
	return m.ValidateConfigResponse, m.ValidateConfigError
}

// ExecuteCLIStreaming mocks streaming CLI execution
func (m *MockWarpgateClient) ExecuteCLIStreaming(_ context.Context, callback OutputCallback, args ...string) (string, error) {
	m.LastExecuteCLIArgs = args
	// Simulate streaming by calling callback for each line
	for _, line := range m.ExecuteCLIStreamingLines {
		if callback != nil {
			callback(line)
		}
	}
	return m.ExecuteCLIResponse, m.ExecuteCLIError
}

// WarpgateBuildStreaming mocks streaming build command
func (m *MockWarpgateClient) WarpgateBuildStreaming(_ context.Context, template string, opts BuildOptions, callback OutputCallback) (string, error) {
	m.LastBuildTemplate = template
	m.LastBuildOptions = opts
	m.LastBuildCallback = callback
	// Simulate streaming by calling callback for each line
	for _, line := range m.BuildStreamingLines {
		if callback != nil {
			callback(line)
		}
	}
	return m.BuildResponse, m.BuildError
}

// RegistryDelete mocks the registry delete command
func (m *MockWarpgateClient) RegistryDelete(_ context.Context, opts RegistryDeleteOptions) (string, error) {
	m.LastRegistryDeleteOptions = opts
	return m.RegistryDeleteResponse, m.RegistryDeleteError
}

// RegistryCopy mocks the registry copy command
func (m *MockWarpgateClient) RegistryCopy(_ context.Context, opts RegistryCopyOptions) (string, error) {
	m.LastRegistryCopyOptions = opts
	return m.RegistryCopyResponse, m.RegistryCopyError
}
