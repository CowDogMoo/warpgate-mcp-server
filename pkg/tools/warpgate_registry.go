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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		registry, ok := request.Params.Arguments["registry"].(string)
		if !ok || registry == "" {
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
			Name:     name,
			Registry: registry,
		}

		if namespace, ok := request.Params.Arguments["namespace"].(string); ok {
			opts.Namespace = namespace
		}

		if authFile, ok := request.Params.Arguments["auth_file"].(string); ok {
			opts.AuthFile = authFile
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		registry, ok := request.Params.Arguments["registry"].(string)
		if !ok || registry == "" {
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
			Name:     name,
			Registry: registry,
		}

		if tags, ok := request.Params.Arguments["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if t, ok := tag.(string); ok {
					opts.Tags = append(opts.Tags, t)
				}
			}
		}

		if namespace, ok := request.Params.Arguments["namespace"].(string); ok {
			opts.Namespace = namespace
		}

		if authFile, ok := request.Params.Arguments["auth_file"].(string); ok {
			opts.AuthFile = authFile
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
