// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterResources registers all available resources with the MCP server
func RegisterResources(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	warpgateConfigResource(s, logger, warpgatePath)
	templateSchemaResource(s, logger, warpgatePath)
	exampleTemplateResource(s, logger, warpgatePath)
}

// warpgateConfigResource exposes the warpgate configuration
func warpgateConfigResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://config",
		Name:        "Warpgate Configuration",
		Description: "Global warpgate configuration from ~/.config/warpgate/config.yaml",
		MIMEType:    "application/yaml",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		config, err := wg.GetWarpgateConfig()
		if err != nil {
			logger.Errorf("Failed to get warpgate config: %v", err)
			return nil, fmt.Errorf("failed to get warpgate config: %w", err)
		}

		// Convert config to JSON for better readability
		configJSON, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}

		return []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(configJSON),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

// templateSchemaResource exposes the warpgate template JSON schema
func templateSchemaResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://schema/template",
		Name:        "Warpgate Template Schema",
		Description: "JSON schema for warpgate.yaml template files",
		MIMEType:    "application/json",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		// Fetch the schema from the warpgate repository
		schemaURL := "https://raw.githubusercontent.com/cowdogmoo/warpgate/main/schema/warpgate-template.json"

		resp, err := http.Get(schemaURL)
		if err != nil {
			logger.Errorf("Failed to fetch schema: %v", err)
			return nil, fmt.Errorf("failed to fetch schema: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch schema: HTTP %d", resp.StatusCode)
		}

		var schema interface{}
		if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
			return nil, fmt.Errorf("failed to decode schema: %w", err)
		}

		schemaJSON, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}

		return []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(schemaJSON),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

// exampleTemplateResource exposes the example warpgate.yaml
func exampleTemplateResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://examples/template",
		Name:        "Example Warpgate Template",
		Description: "Example warpgate.yaml demonstrating the template format",
		MIMEType:    "application/yaml",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		// Try to read the example file from the MCP server repository
		examplePath := filepath.Join(warpgatePath, "..", "warpgate-mcp-server", "examples", "warpgate.yaml")

		// If that doesn't exist, try relative to current directory
		if _, err := os.Stat(examplePath); os.IsNotExist(err) {
			examplePath = "examples/warpgate.yaml"
		}

		content, err := os.ReadFile(examplePath)
		if err != nil {
			logger.Warnf("Failed to read example template: %v", err)
			// Return a minimal example if file doesn't exist
			content = []byte(`---
# Example warpgate template
metadata:
  name: example-template
  version: 1.0.0
  description: Example warpgate template

name: example-app
version: latest

base:
  image: ubuntu:22.04

provisioners:
  - type: shell
    inline:
      - apt-get update
      - apt-get install -y curl

targets:
  - type: container
    registry: ghcr.io/myorg
    tags:
      - latest
    platforms:
      - linux/amd64
      - linux/arm64
    push: false
`)
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
