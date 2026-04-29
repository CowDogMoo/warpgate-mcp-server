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

func searchTemplates(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "search_templates",
		Description: "Search warpgate templates by substring across name, description, author, and tags.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search term (case-insensitive, matched against name/description/author/tags)",
				},
			},
			Required: []string{"query"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := argString(request.Params.Arguments, "query", "")
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		results, err := wg.SearchTemplates(ctx, query)
		if err != nil {
			logger.Errorf("search_templates: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		body, err := json.MarshalIndent(map[string]interface{}{
			"query":   query,
			"count":   len(results),
			"results": results,
		}, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal: %v", err)), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}

	s.AddTool(tool, handler)
}
