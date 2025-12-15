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

func initTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "init_template",
		Description: "Initialize a new warpgate template with scaffolding",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Template name (e.g., 'my-awesome-template')",
				},
				"from_template": map[string]interface{}{
					"type":        "string",
					"description": "Fork from existing template (optional)",
				},
				"output_dir": map[string]interface{}{
					"type":        "string",
					"description": "Output directory (default: current directory)",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse arguments
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}

		fromTemplate := ""
		if ft, ok := request.Params.Arguments["from_template"].(string); ok {
			fromTemplate = ft
		}

		outputDir := ""
		if od, ok := request.Params.Arguments["output_dir"].(string); ok {
			outputDir = od
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		output, err := wg.InitTemplate(name, fromTemplate, outputDir)
		if err != nil {
			logger.Errorf("Failed to initialize template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize template: %v\n%s", err, output)), nil
		}

		logger.Infof("Template %s initialized successfully", name)
		return mcp.NewToolResultText(fmt.Sprintf("Template initialized successfully:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
