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

func listTemplates(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "list_templates",
		Description: "List available warpgate templates from all configured sources",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Filter by source (all, local, git, or specific repo name)",
					"default":     "all",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		// Parse source parameter
		source := "all"
		if request.Params.Arguments != nil {
			if sourceArg, ok := request.Params.Arguments["source"].(string); ok {
				source = sourceArg
			}
		}

		var templates []client.TemplateInfo
		if source == "all" {
			templates, err = wg.ListTemplates()
		} else {
			templates, err = wg.ListTemplatesFromSource(source)
		}

		if err != nil {
			logger.Errorf("Failed to list templates: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list templates: %v", err)), nil
		}

		result := map[string]interface{}{
			"templates": templates,
			"count":     len(templates),
			"source":    source,
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}

	s.AddTool(tool, handler)
}
