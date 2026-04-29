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

func showConfig(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "show_config",
		Description: "Show the current resolved warpgate configuration (defaults + config file + env + flags).",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		out, err := wg.ConfigShow(ctx)
		if err != nil {
			logger.Errorf("show_config: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}

func getConfig(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "get_config",
		Description: "Get a specific warpgate config value by dotted key (e.g. 'log.level', 'aws.region', 'registry.default').",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Dotted config key",
				},
			},
			Required: []string{"key"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key := argString(request.Params.Arguments, "key", "")
		if key == "" {
			return mcp.NewToolResultError("key is required"), nil
		}
		out, err := wg.ConfigGet(ctx, key)
		if err != nil {
			logger.Errorf("get_config: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}

func setConfig(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "set_config",
		Description: "Set a warpgate config value by dotted key (e.g. 'log.level=debug'). Writes to ~/.config/warpgate/config.yaml.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Dotted config key",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "New value",
				},
			},
			Required: []string{"key", "value"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key := argString(request.Params.Arguments, "key", "")
		value := argString(request.Params.Arguments, "value", "")
		if key == "" {
			return mcp.NewToolResultError("key is required"), nil
		}
		out, err := wg.ConfigSet(ctx, key, value)
		if err != nil {
			logger.Errorf("set_config: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
