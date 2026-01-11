// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

// Package client provides the warpgate CLI client interface and implementations.
package client

// WarpgateClientInterface defines the interface for warpgate client operations
// This allows for easy mocking in tests
type WarpgateClientInterface interface {
	// CLI detection
	IsCLIAvailable() bool
	GetCLIVersion() string
	GetBinaryPath() string
	GetRepoPath() string

	// CLI execution
	ExecuteCLI(args ...string) (string, error)
	ExecuteCLIWithWorkdir(workdir string, args ...string) (string, error)
	ExecuteCLIStreaming(callback OutputCallback, args ...string) (string, error)

	// Build operations
	WarpgateBuild(template string, opts BuildOptions) (string, error)
	WarpgateBuildStreaming(template string, opts BuildOptions, callback OutputCallback) (string, error)
	WarpgateValidate(configPath string, syntaxOnly bool) (string, error)
	WarpgateInit(name string, opts InitOptions) (string, error)

	// Template operations
	WarpgateTemplatesList(source, format string) (string, error)
	WarpgateTemplatesInfo(template string) (string, error)
	WarpgateTemplatesAdd(source string, name string) (string, error)
	WarpgateTemplatesRemove(nameOrPath string) (string, error)

	// Manifest operations
	WarpgateManifestsCreate(name string, images []string, push bool) (string, error)
	WarpgateManifestsPush(name string, purge bool) (string, error)
	WarpgateManifestsList(opts ManifestsListOptions) (string, error)
	WarpgateManifestsInspect(opts ManifestsInspectOptions) (string, error)

	// Config operations
	WarpgateConfigGet(key string) (string, error)
	WarpgateConfigSet(key, value string) (string, error)
	WarpgateConfigShow() (string, error)

	// Conversion
	WarpgateConvert(source, output string) (string, error)
	WarpgateValidateConfig(configPath string) (string, error)

	// Registry operations
	RegistryDelete(opts RegistryDeleteOptions) (string, error)
	RegistryCopy(opts RegistryCopyOptions) (string, error)

	// Task operations
	ExecuteTask(taskName string, args map[string]string) (string, error)
	ListTasks() ([]string, error)
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
	ManifestsPushResponse    string
	ManifestsPushError       error
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
	ExecuteTaskResponse      string
	ExecuteTaskError         error
	ListTasksResponse        []string
	ListTasksError           error

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
	LastManifestsCreateName   string
	LastManifestsCreateImages []string
	LastManifestsCreatePush   bool
	LastManifestsPushName     string
	LastManifestsPushPurge    bool
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
	LastExecuteTaskName       string
	LastExecuteTaskArgs       map[string]string
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
func (m *MockWarpgateClient) ExecuteCLI(args ...string) (string, error) {
	m.LastExecuteCLIArgs = args
	return m.ExecuteCLIResponse, m.ExecuteCLIError
}

// ExecuteCLIWithWorkdir mocks CLI execution with workdir
func (m *MockWarpgateClient) ExecuteCLIWithWorkdir(_ string, args ...string) (string, error) {
	m.LastExecuteCLIArgs = args
	return m.ExecuteCLIResponse, m.ExecuteCLIError
}

// WarpgateBuild mocks the build command
func (m *MockWarpgateClient) WarpgateBuild(template string, opts BuildOptions) (string, error) {
	m.LastBuildTemplate = template
	m.LastBuildOptions = opts
	return m.BuildResponse, m.BuildError
}

// WarpgateValidate mocks the validate command
func (m *MockWarpgateClient) WarpgateValidate(configPath string, _ bool) (string, error) {
	m.LastValidatePath = configPath
	return m.ValidateResponse, m.ValidateError
}

// WarpgateInit mocks the init command
func (m *MockWarpgateClient) WarpgateInit(name string, opts InitOptions) (string, error) {
	m.LastInitName = name
	m.LastInitOptions = opts
	return m.InitResponse, m.InitError
}

// WarpgateTemplatesList mocks the templates list command
func (m *MockWarpgateClient) WarpgateTemplatesList(source, format string) (string, error) {
	m.LastTemplatesListSource = source
	m.LastTemplatesListFormat = format
	return m.TemplatesListResponse, m.TemplatesListError
}

// WarpgateTemplatesInfo mocks the templates info command
func (m *MockWarpgateClient) WarpgateTemplatesInfo(template string) (string, error) {
	m.LastTemplatesInfoTemplate = template
	return m.TemplatesInfoResponse, m.TemplatesInfoError
}

// WarpgateTemplatesAdd mocks the templates add command
func (m *MockWarpgateClient) WarpgateTemplatesAdd(source string, name string) (string, error) {
	m.LastTemplatesAddSource = source
	m.LastTemplatesAddName = name
	return m.TemplatesAddResponse, m.TemplatesAddError
}

// WarpgateTemplatesRemove mocks the templates remove command
func (m *MockWarpgateClient) WarpgateTemplatesRemove(nameOrPath string) (string, error) {
	m.LastTemplatesRemoveName = nameOrPath
	return m.TemplatesRemoveResponse, m.TemplatesRemoveError
}

// WarpgateManifestsCreate mocks the manifests create command
func (m *MockWarpgateClient) WarpgateManifestsCreate(name string, images []string, push bool) (string, error) {
	m.LastManifestsCreateName = name
	m.LastManifestsCreateImages = images
	m.LastManifestsCreatePush = push
	return m.ManifestsCreateResponse, m.ManifestsCreateError
}

// WarpgateManifestsPush mocks the manifests push command
func (m *MockWarpgateClient) WarpgateManifestsPush(name string, purge bool) (string, error) {
	m.LastManifestsPushName = name
	m.LastManifestsPushPurge = purge
	return m.ManifestsPushResponse, m.ManifestsPushError
}

// WarpgateManifestsList mocks the manifests list command
func (m *MockWarpgateClient) WarpgateManifestsList(opts ManifestsListOptions) (string, error) {
	m.LastManifestsListOptions = opts
	return m.ManifestsListResponse, m.ManifestsListError
}

// WarpgateManifestsInspect mocks the manifests inspect command
func (m *MockWarpgateClient) WarpgateManifestsInspect(opts ManifestsInspectOptions) (string, error) {
	m.LastManifestsInspectOpts = opts
	return m.ManifestsInspectResponse, m.ManifestsInspectError
}

// WarpgateConfigGet mocks the config get command
func (m *MockWarpgateClient) WarpgateConfigGet(key string) (string, error) {
	m.LastConfigGetKey = key
	return m.ConfigGetResponse, m.ConfigGetError
}

// WarpgateConfigSet mocks the config set command
func (m *MockWarpgateClient) WarpgateConfigSet(key, value string) (string, error) {
	m.LastConfigSetKey = key
	m.LastConfigSetValue = value
	return m.ConfigSetResponse, m.ConfigSetError
}

// WarpgateConfigShow mocks the config show command
func (m *MockWarpgateClient) WarpgateConfigShow() (string, error) {
	return m.ConfigShowResponse, m.ConfigShowError
}

// WarpgateConvert mocks the convert command
func (m *MockWarpgateClient) WarpgateConvert(source, output string) (string, error) {
	m.LastConvertSource = source
	m.LastConvertOutput = output
	return m.ConvertResponse, m.ConvertError
}

// WarpgateValidateConfig mocks the validate config command
func (m *MockWarpgateClient) WarpgateValidateConfig(configPath string) (string, error) {
	m.LastValidateConfigPath = configPath
	return m.ValidateConfigResponse, m.ValidateConfigError
}

// ExecuteTask mocks task execution
func (m *MockWarpgateClient) ExecuteTask(taskName string, args map[string]string) (string, error) {
	m.LastExecuteTaskName = taskName
	m.LastExecuteTaskArgs = args
	return m.ExecuteTaskResponse, m.ExecuteTaskError
}

// ListTasks mocks the list tasks command
func (m *MockWarpgateClient) ListTasks() ([]string, error) {
	return m.ListTasksResponse, m.ListTasksError
}

// ExecuteCLIStreaming mocks streaming CLI execution
func (m *MockWarpgateClient) ExecuteCLIStreaming(callback OutputCallback, args ...string) (string, error) {
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
func (m *MockWarpgateClient) WarpgateBuildStreaming(template string, opts BuildOptions, callback OutputCallback) (string, error) {
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
func (m *MockWarpgateClient) RegistryDelete(opts RegistryDeleteOptions) (string, error) {
	m.LastRegistryDeleteOptions = opts
	return m.RegistryDeleteResponse, m.RegistryDeleteError
}

// RegistryCopy mocks the registry copy command
func (m *MockWarpgateClient) RegistryCopy(opts RegistryCopyOptions) (string, error) {
	m.LastRegistryCopyOptions = opts
	return m.RegistryCopyResponse, m.RegistryCopyError
}
