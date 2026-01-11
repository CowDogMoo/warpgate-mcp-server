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

func warpgateConfigGet(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_config_get",
		Description: "Get warpgate configuration values. Returns specific key value or all configuration if no key specified.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Configuration key to retrieve (e.g., 'registry', 'aws.region'). If not specified, returns all config.",
				},
			},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		key := ""
		if val, ok := request.Params.Arguments["key"].(string); ok {
			key = val
		}

		output, err := wg.WarpgateConfigGet(key)
		if err != nil {
			logger.Errorf("Failed to get config: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get config: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateConfigSet(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_config_set",
		Description: "Set a warpgate configuration value",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Configuration key to set (e.g., 'registry', 'aws.region')",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Value to set for the configuration key",
				},
			},
			Required: []string{"key", "value"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key, ok := request.Params.Arguments["key"].(string)
		if !ok || key == "" {
			return mcp.NewToolResultError("key is required and must be a string"), nil
		}

		value, ok := request.Params.Arguments["value"].(string)
		if !ok {
			return mcp.NewToolResultError("value is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		output, err := wg.WarpgateConfigSet(key, value)
		if err != nil {
			logger.Errorf("Failed to set config: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set config: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully set config %s=%s", key, value)
		result := fmt.Sprintf("Successfully set %s = %s\n%s", key, value, output)
		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}

func warpgateConfigShow(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_config_show",
		Description: "Show the current warpgate configuration including all settings and their sources",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
		},
	}

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		output, err := wg.WarpgateConfigShow()
		if err != nil {
			logger.Errorf("Failed to show config: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to show config: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
