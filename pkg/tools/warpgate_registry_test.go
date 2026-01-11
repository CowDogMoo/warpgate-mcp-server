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
