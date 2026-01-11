// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

// Package resources provides MCP resource handlers for the warpgate-mcp-server.
package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterResources registers all available resources with the MCP server
func RegisterResources(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	templateReadmeResource(s, logger, warpgatePath)
	templateConfigResource(s, logger, warpgatePath)
	warpgateConfigResource(s, logger, warpgatePath)
	warpgateCLIInfoResource(s, logger, warpgatePath)
	warpgateSchemaResource(s, logger, warpgatePath)
}

func templateReadmeResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://template/{name}/readme",
		Name:        "Template README",
		Description: "README.md for a specific template",
		MIMEType:    "text/markdown",
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// This is a template resource, actual handling would need the template name
		// For now, return a list of available templates from the warpgate CLI
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		if !wg.IsCLIAvailable() {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "warpgate://template/{name}/readme",
					MIMEType: "text/markdown",
					Text:     "Warpgate CLI is not available. Please install warpgate >= 1.0.0 to list templates.",
				},
			}, nil
		}

		// Use warpgate CLI to list templates
		templatesOutput, err := wg.WarpgateTemplatesList("", "table")
		if err != nil {
			logger.Errorf("Failed to list templates: %v", err)
			return nil, fmt.Errorf("failed to list templates: %w", err)
		}

		result := fmt.Sprintf(`To access a template README, use URI: warpgate://template/<name>/readme

Available templates (from warpgate templates list):
%s`, templatesOutput)

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "warpgate://template/{name}/readme",
				MIMEType: "text/markdown",
				Text:     result,
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

func templateConfigResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://template/{name}/config",
		Name:        "Template Configuration",
		Description: "The warpgate.yaml configuration for a specific template",
		MIMEType:    "text/yaml",
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		if !wg.IsCLIAvailable() {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "warpgate://template/{name}/config",
					MIMEType: "text/yaml",
					Text:     "Warpgate CLI is not available. Please install warpgate >= 1.0.0 to access template configs.",
				},
			}, nil
		}

		// Use warpgate CLI to get template info which includes the config
		templatesOutput, err := wg.WarpgateTemplatesInfo("")
		if err != nil {
			// Provide usage information since URI template parameter extraction is limited
			result := fmt.Sprintf(`To view a template configuration, use 'warpgate_templates_info' tool with the template name.

URI pattern: warpgate://template/{name}/config

The warpgate CLI 'templates info' command returns the full template configuration
including metadata, provisioners, targets, and computed values.

Example templates (use warpgate_templates_list to see all):
- attack-box
- sliver
- atomic-red-team

Error: %v`, err)
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "warpgate://template/{name}/config",
					MIMEType: "text/yaml",
					Text:     result,
				},
			}, nil
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "warpgate://template/{name}/config",
				MIMEType: "text/yaml",
				Text:     templatesOutput,
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

func warpgateConfigResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://config",
		Name:        "Warpgate Configuration",
		Description: "Current warpgate CLI configuration settings",
		MIMEType:    "text/plain",
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		if !wg.IsCLIAvailable() {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "warpgate://config",
					MIMEType: "text/plain",
					Text:     "Warpgate CLI is not available. Please install warpgate >= 1.0.0",
				},
			}, nil
		}

		output, err := wg.WarpgateConfigShow()
		if err != nil {
			logger.Errorf("Failed to get config: %v", err)
			return nil, fmt.Errorf("failed to get warpgate config: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "warpgate://config",
				MIMEType: "text/plain",
				Text:     output,
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

func warpgateCLIInfoResource(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	resource := mcp.Resource{
		URI:         "warpgate://cli-info",
		Name:        "Warpgate CLI Information",
		Description: "Information about the detected warpgate CLI binary",
		MIMEType:    "application/json",
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return nil, fmt.Errorf("failed to create Warpgate client: %w", err)
		}

		info := map[string]interface{}{
			"cli_available":   wg.IsCLIAvailable(),
			"cli_version":     wg.GetCLIVersion(),
			"binary_path":     wg.GetBinaryPath(),
			"repo_path":       wg.GetRepoPath(),
			"minimum_version": client.MinimumWarpgateVersion,
		}

		resultJSON, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "warpgate://cli-info",
				MIMEType: "application/json",
				Text:     string(resultJSON),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}

func warpgateSchemaResource(s *server.MCPServer, _ *logging.Logger, _ string) {
	resource := mcp.Resource{
		URI:         "warpgate://schema/template",
		Name:        "Warpgate Template Schema",
		Description: "JSON Schema for validating warpgate.yaml template configuration files",
		MIMEType:    "application/json",
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Warpgate template schema based on the warpgate CLI spec
		schema := map[string]interface{}{
			"$schema":     "http://json-schema.org/draft-07/schema#",
			"title":       "Warpgate Template Configuration",
			"description": "Schema for warpgate.yaml template configuration files",
			"type":        "object",
			"required":    []string{"name"},
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Template name (must match directory name)",
					"pattern":     "^[a-z0-9][a-z0-9-]*[a-z0-9]$",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Human-readable description of what this template creates",
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Template version (semantic versioning recommended)",
					"pattern":     "^v?\\d+\\.\\d+\\.\\d+",
				},
				"maintainers": map[string]interface{}{
					"type":        "array",
					"description": "List of template maintainers",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"base_image": map[string]interface{}{
					"type":        "object",
					"description": "Base container image configuration",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Base image name (e.g., ubuntu, alpine)",
						},
						"tag": map[string]interface{}{
							"type":        "string",
							"description": "Base image tag (e.g., 22.04, latest)",
						},
					},
				},
				"targets": map[string]interface{}{
					"type":        "object",
					"description": "Build target configurations",
					"properties": map[string]interface{}{
						"container": map[string]interface{}{
							"type":        "object",
							"description": "Container build target",
							"properties": map[string]interface{}{
								"enabled": map[string]interface{}{
									"type": "boolean",
								},
								"architectures": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "string",
										"enum": []string{"amd64", "arm64"},
									},
								},
							},
						},
						"ami": map[string]interface{}{
							"type":        "object",
							"description": "AWS AMI build target",
							"properties": map[string]interface{}{
								"enabled": map[string]interface{}{
									"type": "boolean",
								},
								"region": map[string]interface{}{
									"type": "string",
								},
								"instance_type": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"provisioners": map[string]interface{}{
					"type":        "array",
					"description": "List of provisioning steps",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{"shell", "file", "ansible"},
							},
							"name": map[string]interface{}{
								"type": "string",
							},
							"inline": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"script": map[string]interface{}{
								"type": "string",
							},
							"source": map[string]interface{}{
								"type": "string",
							},
							"destination": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"variables": map[string]interface{}{
					"type":        "object",
					"description": "Template variables with default values",
					"additionalProperties": map[string]interface{}{
						"oneOf": []map[string]interface{}{
							{"type": "string"},
							{"type": "number"},
							{"type": "boolean"},
						},
					},
				},
			},
		}

		resultJSON, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "warpgate://schema/template",
				MIMEType: "application/json",
				Text:     string(resultJSON),
			},
		}, nil
	}

	s.AddResource(resource, handler)
}
