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

func removeTemplateSource(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "remove_template_source",
		Description: "Unregister a template source by repository name (e.g. 'official') or local path.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path_or_name": map[string]interface{}{
					"type":        "string",
					"description": "Repository name or local path",
				},
			},
			Required: []string{"path_or_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pathOrName := argString(request.Params.Arguments, "path_or_name", "")
		if pathOrName == "" {
			return mcp.NewToolResultError("path_or_name is required"), nil
		}

		out, err := wg.RemoveTemplateSource(ctx, pathOrName)
		if err != nil {
			logger.Errorf("remove_template_source: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
