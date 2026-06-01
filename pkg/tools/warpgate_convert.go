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

func warpgateConvert(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_convert",
		Description: "Convert a Packer template to warpgate YAML format. Helps migrate existing Packer infrastructure to the new warpgate configuration system.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Path to the Packer template file or directory to convert",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output path for the generated warpgate.yaml (default: stdout or source directory)",
				},
			},
			Required: []string{"source"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := request.GetString("source", "")
		if source == "" {
			return mcp.NewToolResultError("source is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 3.0.0"), nil
		}

		output := request.GetString("output", "")

		result, err := wg.WarpgateConvert(ctx, source, output)
		if err != nil {
			logger.Errorf("Failed to convert template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert template: %v\n%s", err, result)), nil
		}

		logger.Infof("Successfully converted template: %s", source)
		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}
