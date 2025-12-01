// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func getTemplateInfo(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "get_template_info",
		Description: "Get detailed information about a specific Packer template",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the template (e.g., attack-box, sliver, atomic-red-team)",
				},
			},
			Required: []string{"template_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateName, ok := request.Params.Arguments["template_name"].(string)
		if !ok {
			return mcp.NewToolResultError("template_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		info, err := wg.GetTemplateInfo(templateName)
		if err != nil {
			logger.Errorf("Failed to get template info: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get template info: %v", err)), nil
		}

		resultJSON, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}

	s.AddTool(tool, handler)
}
