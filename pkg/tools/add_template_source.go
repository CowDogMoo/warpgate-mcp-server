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

func addTemplateSource(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "add_template_source",
		Description: "Register a new template source. Accepts a git URL or a local directory path. For git URLs, an optional 'name' overrides the auto-derived name; ignored for local paths.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url_or_path": map[string]interface{}{
					"type":        "string",
					"description": "Git URL (https://github.com/user/templates.git) or local directory path",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Optional custom name for git repos",
				},
			},
			Required: []string{"url_or_path"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		urlOrPath := argString(request.Params.Arguments, "url_or_path", "")
		if urlOrPath == "" {
			return mcp.NewToolResultError("url_or_path is required"), nil
		}
		name := argString(request.Params.Arguments, "name", "")

		out, err := wg.AddTemplateSource(ctx, urlOrPath, name)
		if err != nil {
			logger.Errorf("add_template_source: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
