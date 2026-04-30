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

func warpgateManifestsCreate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_manifests_create",
		Description: "Create a multi-architecture manifest from digest files generated during separate architecture builds",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the multi-arch manifest (e.g., 'registry/image:tag')",
				},
				"images": map[string]interface{}{
					"type":        "array",
					"description": "List of image references or digest files to include in the manifest",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"push": map[string]interface{}{
					"type":        "boolean",
					"description": "Push the manifest to the registry after creation",
				},
			},
			Required: []string{"name", "images"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		images := request.GetStringSlice("images", nil)
		if len(images) == 0 {
			return mcp.NewToolResultError("images is required and must be a non-empty array"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		push := request.GetBool("push", false)

		output, err := wg.WarpgateManifestsCreate(ctx, name, images, push)
		if err != nil {
			logger.Errorf("Failed to create manifest: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create manifest: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully created manifest: %s", name)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateManifestsPush(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_manifests_push",
		Description: "Push a multi-architecture manifest to the container registry",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the manifest to push (e.g., 'registry/image:tag')",
				},
				"purge": map[string]interface{}{
					"type":        "boolean",
					"description": "Purge local manifest after pushing",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		purge := request.GetBool("purge", false)

		output, err := wg.WarpgateManifestsPush(ctx, name, purge)
		if err != nil {
			logger.Errorf("Failed to push manifest: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to push manifest: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully pushed manifest: %s", name)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
