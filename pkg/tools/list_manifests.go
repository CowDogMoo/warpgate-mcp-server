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

func listManifests(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "list_manifests",
		Description: "List available manifest tags for an image in a registry.",
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

		out, err := wg.ListManifests(ctx, name, registry, argString(args, "namespace", ""))
		if err != nil {
			logger.Errorf("list_manifests: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
