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

func listTasks(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "list_tasks",
		Description: "List all available Taskfile tasks in the Warpgate repository",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		tasks, err := wg.ListTasks()
		if err != nil {
			logger.Errorf("Failed to list tasks: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks: %v", err)), nil
		}

		result := map[string]interface{}{
			"tasks":     tasks,
			"count":     len(tasks),
			"repo_path": wg.GetRepoPath(),
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}

	s.AddTool(tool, handler)
}

func runTask(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "run_task",
		Description: "Run a specific Taskfile task with optional arguments",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"task_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the task to run",
				},
				"args": map[string]interface{}{
					"type":        "object",
					"description": "Arguments to pass to the task as key-value pairs",
				},
			},
			Required: []string{"task_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskName, ok := request.Params.Arguments["task_name"].(string)
		if !ok {
			return mcp.NewToolResultError("task_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		taskArgs := make(map[string]string)
		if argsMap, ok := request.Params.Arguments["args"].(map[string]interface{}); ok {
			for k, v := range argsMap {
				taskArgs[k] = fmt.Sprintf("%v", v)
			}
		}

		output, err := wg.ExecuteTask(taskName, taskArgs)
		if err != nil {
			logger.Errorf("Failed to run task: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to run task: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func runImageBuilder(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "run_image_builder",
		Description: "Run the GitHub Actions image-builder workflow locally using act",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Specific template to build (optional, if not provided builds all)",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		taskArgs := make(map[string]string)
		if template, ok := request.Params.Arguments["template"].(string); ok && template != "" {
			taskArgs["TEMPLATE"] = template
		}

		output, err := wg.ExecuteTask("run-image-builder-action", taskArgs)
		if err != nil {
			logger.Errorf("Failed to run image builder: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to run image builder: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
