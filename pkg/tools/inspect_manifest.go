// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func inspectManifest(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "inspect_manifest",
		Description: "Inspect a multi-architecture manifest in a registry — shows architectures, digests, sizes, and annotations.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"image_name": map[string]interface{}{
					"type":        "string",
					"description": "Image name",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Image namespace/organization",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Tags to inspect (default: ['latest'])",
				},
			},
			Required: []string{"image_name", "registry"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		name := argString(args, "image_name", "")
		registry := argString(args, "registry", "")
		if name == "" || registry == "" {
			return mcp.NewToolResultError("image_name and registry are required"), nil
		}

		out, err := wg.InspectManifest(ctx, name, registry, argString(args, "namespace", ""), argStringSlice(args, "tags"))
		if err != nil {
			logger.Errorf("inspect_manifest: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
