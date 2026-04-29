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

func listTemplates(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "list_templates",
		Description: "List available warpgate templates from configured sources. Returns name, description, version, repository, path, tags, and author for each template.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Filter by source: 'all' (default), 'local', 'git', or a specific repo name",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := argString(request.Params.Arguments, "source", client.SourceAll)

		templates, err := wg.ListTemplates(ctx, source)
		if err != nil {
			logger.Errorf("list_templates: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		body, err := json.MarshalIndent(map[string]interface{}{
			"source":    source,
			"count":     len(templates),
			"templates": templates,
		}, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal: %v", err)), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}

	s.AddTool(tool, handler)
}
