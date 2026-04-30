// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"testing"
)

func TestDetectRegistryTool(t *testing.T) {
	// This test checks if the function works without panicking
	// It may or may not find a tool depending on the test environment
	_, err := DetectRegistryTool()

	// We don't assert on the result because it depends on the environment
	// Just ensure the function runs without panicking
	if err != nil {
		// Expected if no tools are installed
		t.Logf("No registry tool found (expected in some environments): %v", err)
	}
}

func TestBuildImageRef(t *testing.T) {
	tests := []struct {
		name      string
		registry  string
		namespace string
		imageName string
		tag       string
		expected  string
	}{
		{
			name:      "full reference",
			registry:  "ghcr.io",
			namespace: "cowdogmoo",
			imageName: "attack-box",
			tag:       "latest",
			expected:  "ghcr.io/cowdogmoo/attack-box:latest",
		},
		{
			name:      "no namespace",
			registry:  "docker.io",
			namespace: "",
			imageName: "nginx",
			tag:       "1.21",
			expected:  "docker.io/nginx:1.21",
		},
		{
			name:      "no tag",
			registry:  "ghcr.io",
			namespace: "myorg",
			imageName: "myimage",
			tag:       "",
			expected:  "ghcr.io/myorg/myimage",
		},
		{
			name:      "no registry",
			registry:  "",
			namespace: "myorg",
			imageName: "myimage",
			tag:       "v1.0",
			expected:  "myorg/myimage:v1.0",
		},
		{
			name:      "only image name",
			registry:  "",
			namespace: "",
			imageName: "myimage",
			tag:       "",
			expected:  "myimage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildImageRef(tt.registry, tt.namespace, tt.imageName, tt.tag)
			if result != tt.expected {
				t.Errorf("buildImageRef() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRegistryDeleteOptions(t *testing.T) {
	opts := RegistryDeleteOptions{
		Name:      "test-image",
		Registry:  "ghcr.io",
		Namespace: "cowdogmoo",
		Tags:      []string{"v1.0", "latest"},
		AuthFile:  "/path/to/auth.json",
		DryRun:    true,
	}

	if opts.Name != "test-image" {
		t.Errorf("Expected Name to be 'test-image', got %s", opts.Name)
	}
	if opts.Registry != "ghcr.io" {
		t.Errorf("Expected Registry to be 'ghcr.io', got %s", opts.Registry)
	}
	if len(opts.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(opts.Tags))
	}
	if !opts.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

func TestRegistryCopyOptions(t *testing.T) {
	opts := RegistryCopyOptions{
		SourceImage:     "ghcr.io/cowdogmoo/attack-box:latest",
		DestImage:       "docker.io/myorg/attack-box:v1.0",
		SourceAuth:      "/path/to/source-auth.json",
		DestAuth:        "/path/to/dest-auth.json",
		AllTags:         true,
		PreserveDigests: true,
	}

	if opts.SourceImage != "ghcr.io/cowdogmoo/attack-box:latest" {
		t.Errorf("Expected SourceImage to be 'ghcr.io/cowdogmoo/attack-box:latest', got %s", opts.SourceImage)
	}
	if opts.DestImage != "docker.io/myorg/attack-box:v1.0" {
		t.Errorf("Expected DestImage to be 'docker.io/myorg/attack-box:v1.0', got %s", opts.DestImage)
	}
	if !opts.AllTags {
		t.Error("Expected AllTags to be true")
	}
	if !opts.PreserveDigests {
		t.Error("Expected PreserveDigests to be true")
	}
}

func TestMockWarpgateClientStreaming(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.ExecuteCLIStreamingLines = []string{
		"Step 1: Preparing build",
		"Step 2: Building image",
		"Step 3: Complete",
	}
	mock.ExecuteCLIResponse = "Build successful"

	var receivedLines []string
	callback := func(line string) {
		receivedLines = append(receivedLines, line)
	}

	result, err := mock.ExecuteCLIStreaming(context.Background(), callback, "build", "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Build successful" {
		t.Errorf("Expected 'Build successful', got %s", result)
	}

	if len(receivedLines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(receivedLines))
	}

	if receivedLines[0] != "Step 1: Preparing build" {
		t.Errorf("Expected first line to be 'Step 1: Preparing build', got %s", receivedLines[0])
	}
}

func TestMockWarpgateClientBuildStreaming(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.BuildStreamingLines = []string{
		"==> Building container...",
		"==> Step 1/5",
		"==> Step 2/5",
		"==> Step 3/5",
		"==> Step 4/5",
		"==> Step 5/5",
		"==> Build complete!",
	}
	mock.BuildResponse = "Image built: ghcr.io/test/image:latest"

	var receivedLines []string
	callback := func(line string) {
		receivedLines = append(receivedLines, line)
	}

	opts := BuildOptions{
		Target: "container",
		Push:   true,
	}

	result, err := mock.WarpgateBuildStreaming(context.Background(), "test-template", opts, callback)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Image built: ghcr.io/test/image:latest" {
		t.Errorf("Unexpected result: %s", result)
	}

	if len(receivedLines) != 7 {
		t.Errorf("Expected 7 lines, got %d", len(receivedLines))
	}

	if mock.LastBuildTemplate != "test-template" {
		t.Errorf("Expected template 'test-template', got %s", mock.LastBuildTemplate)
	}

	if mock.LastBuildOptions.Target != "container" {
		t.Errorf("Expected target 'container', got %s", mock.LastBuildOptions.Target)
	}
}

func TestMockWarpgateClientRegistryDelete(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.RegistryDeleteResponse = "Deleted: ghcr.io/test/image:v1.0\nDeleted: ghcr.io/test/image:latest"

	opts := RegistryDeleteOptions{
		Name:     "image",
		Registry: "ghcr.io/test",
		Tags:     []string{"v1.0", "latest"},
		DryRun:   false,
	}

	result, err := mock.RegistryDelete(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != mock.RegistryDeleteResponse {
		t.Errorf("Unexpected result: %s", result)
	}

	if mock.LastRegistryDeleteOptions.Name != "image" {
		t.Errorf("Expected name 'image', got %s", mock.LastRegistryDeleteOptions.Name)
	}

	if len(mock.LastRegistryDeleteOptions.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(mock.LastRegistryDeleteOptions.Tags))
	}
}

func TestMockWarpgateClientRegistryCopy(t *testing.T) {
	mock := NewMockWarpgateClient()
	mock.RegistryCopyResponse = "Successfully copied ghcr.io/src/image:latest to docker.io/dest/image:v1.0"

	opts := RegistryCopyOptions{
		SourceImage:     "ghcr.io/src/image:latest",
		DestImage:       "docker.io/dest/image:v1.0",
		AllTags:         false,
		PreserveDigests: true,
	}

	result, err := mock.RegistryCopy(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != mock.RegistryCopyResponse {
		t.Errorf("Unexpected result: %s", result)
	}

	if mock.LastRegistryCopyOptions.SourceImage != "ghcr.io/src/image:latest" {
		t.Errorf("Expected source 'ghcr.io/src/image:latest', got %s", mock.LastRegistryCopyOptions.SourceImage)
	}

	if mock.LastRegistryCopyOptions.DestImage != "docker.io/dest/image:v1.0" {
		t.Errorf("Expected dest 'docker.io/dest/image:v1.0', got %s", mock.LastRegistryCopyOptions.DestImage)
	}

	if !mock.LastRegistryCopyOptions.PreserveDigests {
		t.Error("Expected PreserveDigests to be true")
	}
}

func TestSplitLinesOrCarriageReturn(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected []string
	}{
		{
			name:     "newline separated",
			data:     "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "carriage return separated",
			data:     "progress 10%\rprogress 50%\rprogress 100%",
			expected: []string{"progress 10%", "progress 50%", "progress 100%"},
		},
		{
			name:     "mixed separators",
			data:     "line1\nline2\rprogress\nline4",
			expected: []string{"line1", "line2", "progress", "line4"},
		},
		{
			name:     "single line",
			data:     "single line only",
			expected: []string{"single line only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string
			data := []byte(tt.data)
			offset := 0

			for offset < len(data) {
				advance, token, err := splitLinesOrCarriageReturn(data[offset:], offset+len(data[offset:]) >= len(data))
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if advance == 0 {
					break
				}
				if len(token) > 0 {
					result = append(result, string(token))
				}
				offset += advance
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}
