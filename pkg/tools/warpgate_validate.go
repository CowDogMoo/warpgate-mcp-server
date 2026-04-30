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

func warpgateValidate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_validate",
		Description: "Validate a warpgate template configuration. Checks YAML syntax, schema compliance, and optionally verifies that referenced files exist.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Path to warpgate.yaml config file or template directory",
				},
				"syntax_only": map[string]interface{}{
					"type":        "boolean",
					"description": "Only validate syntax and structure, skip file existence checks",
				},
			},
			Required: []string{"template"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		template := request.GetString("template", "")
		if template == "" {
			return mcp.NewToolResultError("template is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		syntaxOnly := request.GetBool("syntax_only", false)

		output, err := wg.WarpgateValidate(ctx, template, syntaxOnly)
		if err != nil {
			logger.Errorf("Validation failed: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Validation failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Validation passed for template: %s", template)
		result := fmt.Sprintf("Validation successful for: %s\n\n%s", template, output)
		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}
