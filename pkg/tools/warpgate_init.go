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

func warpgateInit(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_init",
		Description: "Initialize a new warpgate template with scaffolding. Creates a warpgate.yaml configuration file, README.md, and scripts directory.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the new template",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output directory for the template (default: current directory)",
				},
				"from": map[string]interface{}{
					"type":        "string",
					"description": "Fork from an existing template as a starting point",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		opts := client.InitOptions{}

		if output, ok := request.Params.Arguments["output"].(string); ok {
			opts.OutputDir = output
		}

		if from, ok := request.Params.Arguments["from"].(string); ok {
			opts.FromTemplate = from
		}

		output, err := wg.WarpgateInit(name, opts)
		if err != nil {
			logger.Errorf("Init failed: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Init failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully initialized template: %s", name)

		result := fmt.Sprintf(`Successfully initialized template '%s'

%s

Next steps:
1. Edit warpgate.yaml to configure your template
2. Add provisioning scripts to the scripts/ directory
3. Validate with: warpgate validate
4. Build with: warpgate build
`, name, output)

		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}
