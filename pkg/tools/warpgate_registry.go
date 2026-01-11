// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func warpgateRegistryList(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_registry_list",
		Description: "List available image tags in a container registry for a given image name",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name (e.g., attack-box, sliver)",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry URL (e.g., ghcr.io/cowdogmoo)",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Optional namespace/organization within the registry",
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to authentication file",
				},
			},
			Required: []string{"name", "registry"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		registry := request.GetString("registry", "")
		if registry == "" {
			return mcp.NewToolResultError("registry is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		opts := client.ManifestsListOptions{
			Name:      name,
			Registry:  registry,
			Namespace: request.GetString("namespace", ""),
			AuthFile:  request.GetString("auth_file", ""),
		}

		output, err := wg.WarpgateManifestsList(opts)
		if err != nil {
			logger.Errorf("Failed to list registry images: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list registry images: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateRegistryInspect(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_registry_inspect",
		Description: "Inspect a container image manifest from a registry, showing architecture details, digests, and annotations",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name (e.g., attack-box, sliver)",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry URL (e.g., ghcr.io/cowdogmoo)",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Image tags to inspect (default: latest)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Optional namespace/organization within the registry",
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to authentication file",
				},
			},
			Required: []string{"name", "registry"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		registry := request.GetString("registry", "")
		if registry == "" {
			return mcp.NewToolResultError("registry is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		opts := client.ManifestsInspectOptions{
			Name:      name,
			Registry:  registry,
			Tags:      request.GetStringSlice("tags", nil),
			Namespace: request.GetString("namespace", ""),
			AuthFile:  request.GetString("auth_file", ""),
		}

		output, err := wg.WarpgateManifestsInspect(opts)
		if err != nil {
			logger.Errorf("Failed to inspect registry image: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to inspect registry image: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateRegistryDelete(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_registry_delete",
		Description: "Delete container images from a registry. Requires skopeo or crane to be installed. Use dry_run to preview deletions before executing.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name (e.g., attack-box, sliver)",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry URL (e.g., ghcr.io/cowdogmoo)",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Image tags to delete",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Optional namespace/organization within the registry",
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to authentication file",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview deletions without actually deleting (default: false)",
				},
			},
			Required: []string{"name", "registry", "tags"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		registry := request.GetString("registry", "")
		if registry == "" {
			return mcp.NewToolResultError("registry is required and must be a string"), nil
		}

		tags := request.GetStringSlice("tags", nil)
		if len(tags) == 0 {
			return mcp.NewToolResultError("tags is required and must be a non-empty array"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		opts := client.RegistryDeleteOptions{
			Name:      name,
			Registry:  registry,
			Tags:      tags,
			Namespace: request.GetString("namespace", ""),
			AuthFile:  request.GetString("auth_file", ""),
			DryRun:    request.GetBool("dry_run", false),
		}

		output, err := wg.RegistryDelete(opts)
		if err != nil {
			logger.Errorf("Failed to delete registry images: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete registry images: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateRegistryCopy(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_registry_copy",
		Description: "Copy container images between registries. Requires skopeo or crane to be installed. Supports multi-architecture images and preserving digests.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source_image": map[string]interface{}{
					"type":        "string",
					"description": "Full source image reference (e.g., ghcr.io/cowdogmoo/attack-box:latest)",
				},
				"dest_image": map[string]interface{}{
					"type":        "string",
					"description": "Full destination image reference (e.g., docker.io/myorg/attack-box:v1.0)",
				},
				"source_auth": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to authentication file for source registry",
				},
				"dest_auth": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to authentication file for destination registry",
				},
				"all_tags": map[string]interface{}{
					"type":        "boolean",
					"description": "Copy all tags from source image (default: false)",
				},
				"preserve_digests": map[string]interface{}{
					"type":        "boolean",
					"description": "Preserve image digests during copy (default: false)",
				},
			},
			Required: []string{"source_image", "dest_image"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sourceImage := request.GetString("source_image", "")
		if sourceImage == "" {
			return mcp.NewToolResultError("source_image is required and must be a string"), nil
		}

		destImage := request.GetString("dest_image", "")
		if destImage == "" {
			return mcp.NewToolResultError("dest_image is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		opts := client.RegistryCopyOptions{
			SourceImage:     sourceImage,
			DestImage:       destImage,
			SourceAuth:      request.GetString("source_auth", ""),
			DestAuth:        request.GetString("dest_auth", ""),
			AllTags:         request.GetBool("all_tags", false),
			PreserveDigests: request.GetBool("preserve_digests", false),
		}

		output, err := wg.RegistryCopy(opts)
		if err != nil {
			logger.Errorf("Failed to copy registry image: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to copy registry image: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully copied image from %s to %s", sourceImage, destImage)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
