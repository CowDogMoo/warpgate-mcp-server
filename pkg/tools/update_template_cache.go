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

func updateTemplateCache(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "update_template_cache",
		Description: "Pull the latest templates from all configured git repositories.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		out, err := wg.UpdateTemplateCache(ctx)
		if err != nil {
			logger.Errorf("update_template_cache: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
