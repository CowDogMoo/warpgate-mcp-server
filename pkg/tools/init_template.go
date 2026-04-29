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

func initTemplate(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "init_template",
		Description: "Scaffold a new warpgate template directory containing warpgate.yaml, README.md, and a scripts/ folder.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Template name (becomes the directory name)",
				},
				"from_template": map[string]interface{}{
					"type":        "string",
					"description": "Optional: fork from an existing template by name",
				},
				"output_dir": map[string]interface{}{
					"type":        "string",
					"description": "Output directory (default: current working directory)",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := argString(request.Params.Arguments, "name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}
		from := argString(request.Params.Arguments, "from_template", "")
		out := argString(request.Params.Arguments, "output_dir", "")

		output, err := wg.InitTemplate(ctx, name, from, out)
		if err != nil {
			logger.Errorf("init_template: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
