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

func warpgateTemplatesList(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_templates_list",
		Description: "List all available warpgate templates from configured sources (local directories, git repositories, official registry)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Filter by source: 'all', 'local', 'git', or specific repository name",
					"default":     "all",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'table', 'json', or 'gha-matrix'",
					"enum":        []string{"table", "json", "gha-matrix"},
					"default":     "table",
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

		source := request.GetString("source", "")
		format := request.GetString("format", "")

		output, err := wg.WarpgateTemplatesList(source, format)
		if err != nil {
			logger.Errorf("Failed to list templates: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list templates: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateTemplatesInfo(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_templates_info",
		Description: "Get detailed information about a specific warpgate template including metadata, provisioners, targets, and dependencies",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Template name to get information about",
				},
			},
			Required: []string{"template"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		output, err := wg.WarpgateTemplatesInfo(template)
		if err != nil {
			logger.Errorf("Failed to get template info: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get template info: %v\n%s", err, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateTemplatesAdd(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_templates_add",
		Description: "Add a template source to the warpgate registry. Sources can be git URLs or local directory paths.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Git URL or local directory path containing templates",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Optional alias/name for the source (auto-generated for git URLs if not specified)",
				},
			},
			Required: []string{"source"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 1.0.0"), nil
		}

		name := request.GetString("name", "")

		output, err := wg.WarpgateTemplatesAdd(source, name)
		if err != nil {
			logger.Errorf("Failed to add template source: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add template source: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully added template source: %s", source)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateTemplatesRemove(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_templates_remove",
		Description: "Remove a template source from the warpgate registry by path or name",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Source name, path, or alias to remove",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
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

		output, err := wg.WarpgateTemplatesRemove(name)
		if err != nil {
			logger.Errorf("Failed to remove template source: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove template source: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully removed template source: %s", name)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
