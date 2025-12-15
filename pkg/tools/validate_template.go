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

func validateTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "validate_template",
		Description: "Validate a warpgate template configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"config": map[string]interface{}{
					"type":        "string",
					"description": "Path to warpgate.yaml or template name",
				},
				"syntax_only": map[string]interface{}{
					"type":        "boolean",
					"description": "Only validate syntax, skip file checks",
					"default":     false,
				},
			},
			Required: []string{"config"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		config, ok := request.Params.Arguments["config"].(string)
		if !ok {
			return mcp.NewToolResultError("config must be a string"), nil
		}

		syntaxOnly := false
		if so, ok := request.Params.Arguments["syntax_only"].(bool); ok {
			syntaxOnly = so
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		output, err := wg.ValidateTemplate(config, syntaxOnly)
		if err != nil {
			logger.Errorf("Failed to validate template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Validation failed: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Template validation successful:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
