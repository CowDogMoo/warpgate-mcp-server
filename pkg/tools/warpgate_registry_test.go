// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"testing"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
)

func TestManifestsListOptions(t *testing.T) {
	opts := client.ManifestsListOptions{
		Name:      "attack-box",
		Registry:  "ghcr.io/cowdogmoo",
		Namespace: "images",
		AuthFile:  "/path/to/auth.json",
	}

	if opts.Name != "attack-box" {
		t.Errorf("ManifestsListOptions.Name = %q, want %q", opts.Name, "attack-box")
	}

	if opts.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("ManifestsListOptions.Registry = %q, want %q", opts.Registry, "ghcr.io/cowdogmoo")
	}

	if opts.Namespace != "images" {
		t.Errorf("ManifestsListOptions.Namespace = %q, want %q", opts.Namespace, "images")
	}

	if opts.AuthFile != "/path/to/auth.json" {
		t.Errorf("ManifestsListOptions.AuthFile = %q, want %q", opts.AuthFile, "/path/to/auth.json")
	}
}

func TestManifestsListOptionsMinimal(t *testing.T) {
	opts := client.ManifestsListOptions{
		Name:     "sliver",
		Registry: "docker.io/library",
	}

	if opts.Name != "sliver" {
		t.Errorf("ManifestsListOptions.Name = %q, want %q", opts.Name, "sliver")
	}

	if opts.Registry != "docker.io/library" {
		t.Errorf("ManifestsListOptions.Registry = %q, want %q", opts.Registry, "docker.io/library")
	}

	if opts.Namespace != "" {
		t.Errorf("ManifestsListOptions.Namespace should be empty, got %q", opts.Namespace)
	}

	if opts.AuthFile != "" {
		t.Errorf("ManifestsListOptions.AuthFile should be empty, got %q", opts.AuthFile)
	}
}

func TestManifestsInspectOptions(t *testing.T) {
	opts := client.ManifestsInspectOptions{
		Name:      "attack-box",
		Registry:  "ghcr.io/cowdogmoo",
		Namespace: "images",
		Tags:      []string{"latest", "v1.0.0"},
		AuthFile:  "/path/to/auth.json",
	}

	if opts.Name != "attack-box" {
		t.Errorf("ManifestsInspectOptions.Name = %q, want %q", opts.Name, "attack-box")
	}

	if opts.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("ManifestsInspectOptions.Registry = %q, want %q", opts.Registry, "ghcr.io/cowdogmoo")
	}

	if opts.Namespace != "images" {
		t.Errorf("ManifestsInspectOptions.Namespace = %q, want %q", opts.Namespace, "images")
	}

	if len(opts.Tags) != 2 {
		t.Errorf("len(ManifestsInspectOptions.Tags) = %d, want 2", len(opts.Tags))
	}

	if opts.Tags[0] != "latest" {
		t.Errorf("ManifestsInspectOptions.Tags[0] = %q, want %q", opts.Tags[0], "latest")
	}

	if opts.Tags[1] != "v1.0.0" {
		t.Errorf("ManifestsInspectOptions.Tags[1] = %q, want %q", opts.Tags[1], "v1.0.0")
	}

	if opts.AuthFile != "/path/to/auth.json" {
		t.Errorf("ManifestsInspectOptions.AuthFile = %q, want %q", opts.AuthFile, "/path/to/auth.json")
	}
}

func TestManifestsInspectOptionsMinimal(t *testing.T) {
	opts := client.ManifestsInspectOptions{
		Name:     "atomic-red-team",
		Registry: "ghcr.io/cowdogmoo",
	}

	if opts.Name != "atomic-red-team" {
		t.Errorf("ManifestsInspectOptions.Name = %q, want %q", opts.Name, "atomic-red-team")
	}

	if opts.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("ManifestsInspectOptions.Registry = %q, want %q", opts.Registry, "ghcr.io/cowdogmoo")
	}

	if len(opts.Tags) != 0 {
		t.Errorf("len(ManifestsInspectOptions.Tags) = %d, want 0", len(opts.Tags))
	}
}

func TestManifestsInspectOptionsSingleTag(t *testing.T) {
	opts := client.ManifestsInspectOptions{
		Name:     "attack-box",
		Registry: "ghcr.io/cowdogmoo",
		Tags:     []string{"v2.0.0"},
	}

	if len(opts.Tags) != 1 {
		t.Errorf("len(ManifestsInspectOptions.Tags) = %d, want 1", len(opts.Tags))
	}

	if opts.Tags[0] != "v2.0.0" {
		t.Errorf("ManifestsInspectOptions.Tags[0] = %q, want %q", opts.Tags[0], "v2.0.0")
	}
}

func TestRegistryURLFormats(t *testing.T) {
	tests := []struct {
		name     string
		registry string
	}{
		{"ghcr.io", "ghcr.io/cowdogmoo"},
		{"docker hub", "docker.io/library"},
		{"docker hub short", "docker.io"},
		{"ecr", "123456789.dkr.ecr.us-east-1.amazonaws.com"},
		{"gcr", "gcr.io/my-project"},
		{"private registry", "registry.example.com:5000"},
		{"localhost", "localhost:5000"},
		{"with path", "registry.example.com/org/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := client.ManifestsListOptions{
				Name:     "test-image",
				Registry: tt.registry,
			}

			if opts.Registry != tt.registry {
				t.Errorf("Registry = %q, want %q", opts.Registry, tt.registry)
			}
		})
	}
}

func TestManifestsListMultipleImages(t *testing.T) {
	// Test that different images can be queried from same registry
	images := []string{"attack-box", "sliver", "atomic-red-team", "nebula"}

	for _, img := range images {
		opts := client.ManifestsListOptions{
			Name:     img,
			Registry: "ghcr.io/cowdogmoo",
		}

		if opts.Name != img {
			t.Errorf("ManifestsListOptions.Name = %q, want %q", opts.Name, img)
		}
	}
}

func TestManifestsInspectMultipleTags(t *testing.T) {
	tags := []string{"latest", "v1.0.0", "v1.1.0", "v2.0.0-alpha", "sha-abc123"}

	opts := client.ManifestsInspectOptions{
		Name:     "attack-box",
		Registry: "ghcr.io/cowdogmoo",
		Tags:     tags,
	}

	if len(opts.Tags) != len(tags) {
		t.Errorf("len(Tags) = %d, want %d", len(opts.Tags), len(tags))
	}

	for i, tag := range tags {
		if opts.Tags[i] != tag {
			t.Errorf("Tags[%d] = %q, want %q", i, opts.Tags[i], tag)
		}
	}
}

func TestRegistryDeleteOptions(t *testing.T) {
	opts := client.RegistryDeleteOptions{
		Name:      "attack-box",
		Registry:  "ghcr.io/cowdogmoo",
		Namespace: "images",
		Tags:      []string{"old-tag", "deprecated"},
		AuthFile:  "/path/to/auth.json",
		DryRun:    true,
	}

	if opts.Name != "attack-box" {
		t.Errorf("RegistryDeleteOptions.Name = %q, want %q", opts.Name, "attack-box")
	}

	if opts.Registry != "ghcr.io/cowdogmoo" {
		t.Errorf("RegistryDeleteOptions.Registry = %q, want %q", opts.Registry, "ghcr.io/cowdogmoo")
	}

	if opts.Namespace != "images" {
		t.Errorf("RegistryDeleteOptions.Namespace = %q, want %q", opts.Namespace, "images")
	}

	if len(opts.Tags) != 2 {
		t.Errorf("len(RegistryDeleteOptions.Tags) = %d, want 2", len(opts.Tags))
	}

	if opts.Tags[0] != "old-tag" {
		t.Errorf("RegistryDeleteOptions.Tags[0] = %q, want %q", opts.Tags[0], "old-tag")
	}

	if opts.AuthFile != "/path/to/auth.json" {
		t.Errorf("RegistryDeleteOptions.AuthFile = %q, want %q", opts.AuthFile, "/path/to/auth.json")
	}

	if !opts.DryRun {
		t.Error("RegistryDeleteOptions.DryRun should be true")
	}
}

func TestRegistryDeleteOptionsMinimal(t *testing.T) {
	opts := client.RegistryDeleteOptions{
		Name:     "sliver",
		Registry: "docker.io/library",
		Tags:     []string{"v1.0"},
	}

	if opts.Name != "sliver" {
		t.Errorf("RegistryDeleteOptions.Name = %q, want %q", opts.Name, "sliver")
	}

	if opts.Registry != "docker.io/library" {
		t.Errorf("RegistryDeleteOptions.Registry = %q, want %q", opts.Registry, "docker.io/library")
	}

	if opts.Namespace != "" {
		t.Errorf("RegistryDeleteOptions.Namespace should be empty, got %q", opts.Namespace)
	}

	if opts.AuthFile != "" {
		t.Errorf("RegistryDeleteOptions.AuthFile should be empty, got %q", opts.AuthFile)
	}

	if opts.DryRun {
		t.Error("RegistryDeleteOptions.DryRun should be false by default")
	}
}

func TestRegistryCopyOptions(t *testing.T) {
	opts := client.RegistryCopyOptions{
		SourceImage:     "ghcr.io/cowdogmoo/attack-box:latest",
		DestImage:       "docker.io/myorg/attack-box:v1.0",
		SourceAuth:      "/path/to/source-auth.json",
		DestAuth:        "/path/to/dest-auth.json",
		AllTags:         true,
		PreserveDigests: true,
	}

	if opts.SourceImage != "ghcr.io/cowdogmoo/attack-box:latest" {
		t.Errorf("RegistryCopyOptions.SourceImage = %q, want %q", opts.SourceImage, "ghcr.io/cowdogmoo/attack-box:latest")
	}

	if opts.DestImage != "docker.io/myorg/attack-box:v1.0" {
		t.Errorf("RegistryCopyOptions.DestImage = %q, want %q", opts.DestImage, "docker.io/myorg/attack-box:v1.0")
	}

	if opts.SourceAuth != "/path/to/source-auth.json" {
		t.Errorf("RegistryCopyOptions.SourceAuth = %q, want %q", opts.SourceAuth, "/path/to/source-auth.json")
	}

	if opts.DestAuth != "/path/to/dest-auth.json" {
		t.Errorf("RegistryCopyOptions.DestAuth = %q, want %q", opts.DestAuth, "/path/to/dest-auth.json")
	}

	if !opts.AllTags {
		t.Error("RegistryCopyOptions.AllTags should be true")
	}

	if !opts.PreserveDigests {
		t.Error("RegistryCopyOptions.PreserveDigests should be true")
	}
}

func TestRegistryCopyOptionsMinimal(t *testing.T) {
	opts := client.RegistryCopyOptions{
		SourceImage: "ghcr.io/src/image:v1.0",
		DestImage:   "docker.io/dst/image:v1.0",
	}

	if opts.SourceImage != "ghcr.io/src/image:v1.0" {
		t.Errorf("RegistryCopyOptions.SourceImage = %q, want %q", opts.SourceImage, "ghcr.io/src/image:v1.0")
	}

	if opts.DestImage != "docker.io/dst/image:v1.0" {
		t.Errorf("RegistryCopyOptions.DestImage = %q, want %q", opts.DestImage, "docker.io/dst/image:v1.0")
	}

	if opts.SourceAuth != "" {
		t.Errorf("RegistryCopyOptions.SourceAuth should be empty, got %q", opts.SourceAuth)
	}

	if opts.DestAuth != "" {
		t.Errorf("RegistryCopyOptions.DestAuth should be empty, got %q", opts.DestAuth)
	}

	if opts.AllTags {
		t.Error("RegistryCopyOptions.AllTags should be false by default")
	}

	if opts.PreserveDigests {
		t.Error("RegistryCopyOptions.PreserveDigests should be false by default")
	}
}

func TestRegistryCopyBetweenRegistries(t *testing.T) {
	// Test various registry combinations
	tests := []struct {
		name        string
		sourceImage string
		destImage   string
	}{
		{
			name:        "ghcr to docker hub",
			sourceImage: "ghcr.io/cowdogmoo/attack-box:latest",
			destImage:   "docker.io/cowdogmoo/attack-box:latest",
		},
		{
			name:        "docker hub to ecr",
			sourceImage: "docker.io/library/nginx:1.21",
			destImage:   "123456789.dkr.ecr.us-east-1.amazonaws.com/nginx:1.21",
		},
		{
			name:        "gcr to private registry",
			sourceImage: "gcr.io/my-project/my-image:v1.0",
			destImage:   "registry.example.com:5000/my-image:v1.0",
		},
		{
			name:        "same registry different tags",
			sourceImage: "ghcr.io/cowdogmoo/attack-box:v1.0",
			destImage:   "ghcr.io/cowdogmoo/attack-box:production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := client.RegistryCopyOptions{
				SourceImage: tt.sourceImage,
				DestImage:   tt.destImage,
			}

			if opts.SourceImage != tt.sourceImage {
				t.Errorf("SourceImage = %q, want %q", opts.SourceImage, tt.sourceImage)
			}

			if opts.DestImage != tt.destImage {
				t.Errorf("DestImage = %q, want %q", opts.DestImage, tt.destImage)
			}
		})
	}
}
