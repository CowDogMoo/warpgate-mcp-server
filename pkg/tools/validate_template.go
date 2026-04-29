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

func validateTemplate(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "validate_template",
		Description: "Validate a warpgate template configuration. Full validation checks file existence; --syntax-only skips file checks.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"config": map[string]interface{}{
					"type":        "string",
					"description": "Path to a warpgate.yaml or a template name",
				},
				"syntax_only": map[string]interface{}{
					"type":        "boolean",
					"description": "Validate structure/syntax only, skip file existence checks",
					"default":     false,
				},
			},
			Required: []string{"config"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		config := argString(request.Params.Arguments, "config", "")
		if config == "" {
			return mcp.NewToolResultError("config is required"), nil
		}
		syntaxOnly := argBool(request.Params.Arguments, "syntax_only", false)

		out, err := wg.ValidateTemplate(ctx, config, syntaxOnly)
		if err != nil {
			logger.Errorf("validate_template: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
