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

func removeTemplateSource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "remove_template_source",
		Description: "Remove a template source from warpgate configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path_or_name": map[string]interface{}{
					"type":        "string",
					"description": "Path or repository name to remove (e.g., 'official' or '/path/to/templates')",
				},
			},
			Required: []string{"path_or_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pathOrName, ok := request.Params.Arguments["path_or_name"].(string)
		if !ok {
			return mcp.NewToolResultError("path_or_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		output, err := wg.RemoveTemplateSource(pathOrName)
		if err != nil {
			logger.Errorf("Failed to remove template source: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove template source: %v\n%s", err, output)), nil
		}

		logger.Infof("Template source removed successfully: %s", pathOrName)
		return mcp.NewToolResultText(fmt.Sprintf("Template source removed successfully:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
