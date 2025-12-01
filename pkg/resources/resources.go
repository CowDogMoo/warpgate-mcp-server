// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterResources registers all available resources with the MCP server
func RegisterResources(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	taskfileResource(s, logger, warpgatePath)
	templateReadmeResource(s, logger, warpgatePath)
}

func taskfileResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://taskfile",
		Name:        "Warpgate Taskfile",
		Description: "The main Taskfile.yaml configuration for Warpgate",
		MIMEType:    "text/yaml",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		taskfilePath := filepath.Join(wg.GetRepoPath(), "Taskfile.yaml")
		content, err := os.ReadFile(taskfilePath)
		if err != nil {
			logger.Errorf("Failed to read Taskfile: %v", err)
			return nil, fmt.Errorf("failed to read Taskfile: %w", err)
		}

		return []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(content),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

func templateReadmeResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://template/{name}/readme",
		Name:        "Template README",
		Description: "README.md for a specific template",
		MIMEType:    "text/markdown",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		// This is a template resource, actual handling would need the template name
		// For now, return a list of available templates
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		templates, err := wg.ListTemplates()
		if err != nil {
			logger.Errorf("Failed to list templates: %v", err)
			return nil, fmt.Errorf("failed to list templates: %w", err)
		}

		result := map[string]interface{}{
			"message":   "To access a template README, use URI: warpgate://template/<name>/readme",
			"available": templates,
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}
