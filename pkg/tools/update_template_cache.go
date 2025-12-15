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

func updateTemplateCache(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "update_template_cache",
		Description: "Update the local cache of templates from all configured git repositories",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		output, err := wg.UpdateTemplateCache()
		if err != nil {
			logger.Errorf("Failed to update template cache: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update template cache: %v\n%s", err, output)), nil
		}

		logger.Infof("Template cache updated successfully")
		return mcp.NewToolResultText(fmt.Sprintf("Template cache updated successfully:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
