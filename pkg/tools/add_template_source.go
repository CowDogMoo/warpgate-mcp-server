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

func addTemplateSource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "add_template_source",
		Description: "Add a template source (git repository or local path) to warpgate configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url_or_path": map[string]interface{}{
					"type":        "string",
					"description": "Git URL (e.g., https://github.com/user/templates.git) or local directory path",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Custom name for git repository (optional, only used for git URLs)",
				},
			},
			Required: []string{"url_or_path"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		urlOrPath, ok := request.Params.Arguments["url_or_path"].(string)
		if !ok {
			return mcp.NewToolResultError("url_or_path must be a string"), nil
		}

		name := ""
		if n, ok := request.Params.Arguments["name"].(string); ok {
			name = n
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		output, err := wg.AddTemplateSource(urlOrPath, name)
		if err != nil {
			logger.Errorf("Failed to add template source: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add template source: %v\n%s", err, output)), nil
		}

		logger.Infof("Template source added successfully: %s", urlOrPath)
		return mcp.NewToolResultText(fmt.Sprintf("Template source added successfully:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
