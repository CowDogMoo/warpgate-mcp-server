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
		Description: "Initialize a Packer template (creates lockfiles and initializes plugins)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the template to initialize",
				},
			},
			Required: []string{"template_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateName, ok := request.Params.Arguments["template_name"].(string)
		if !ok {
			return mcp.NewToolResultError("template_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		taskArgs := map[string]string{
			"TEMPLATE_NAME": templateName,
		}

		output, err := wg.ExecuteTask("template-init", taskArgs)
		if err != nil {
			logger.Errorf("Failed to initialize template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize template: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func validateTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "validate_template",
		Description: "Validate a Packer template for syntax and configuration errors",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the template to validate",
				},
			},
			Required: []string{"template_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateName, ok := request.Params.Arguments["template_name"].(string)
		if !ok {
			return mcp.NewToolResultError("template_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		taskArgs := map[string]string{
			"TEMPLATE_NAME": templateName,
		}

		output, err := wg.ExecuteTask("template-validate", taskArgs)
		if err != nil {
			logger.Errorf("Failed to validate template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to validate template: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func buildTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "build_template",
		Description: "Build a Packer template (creates Docker images or AMIs)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the template to build",
				},
				"only": map[string]interface{}{
					"type":        "string",
					"description": "Build filter (e.g., 'docker.amd64', 'docker.*', 'amazon-ebs.*')",
				},
				"vars": map[string]interface{}{
					"type":        "string",
					"description": "Additional variables in 'key=value key2=value2' format",
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force rebuild even if artifacts exist",
				},
			},
			Required: []string{"template_name"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateName, ok := request.Params.Arguments["template_name"].(string)
		if !ok {
			return mcp.NewToolResultError("template_name must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		taskArgs := map[string]string{
			"TEMPLATE_NAME": templateName,
		}

		if only, ok := request.Params.Arguments["only"].(string); ok && only != "" {
			taskArgs["ONLY"] = only
		}

		if vars, ok := request.Params.Arguments["vars"].(string); ok && vars != "" {
			taskArgs["VARS"] = vars
		}

		if force, ok := request.Params.Arguments["force"].(bool); ok && force {
			taskArgs["FORCE"] = "true"
		}

		output, err := wg.ExecuteTask("template-build", taskArgs)
		if err != nil {
			logger.Errorf("Failed to build template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to build template: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
