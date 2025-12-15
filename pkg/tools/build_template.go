// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func buildTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "build_template",
		Description: "Build a container image or AMI from a warpgate template",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Template name or path to warpgate.yaml",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Override build target (container or ami)",
					"enum":        []string{"container", "ami"},
				},
				"architectures": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Architectures to build (e.g., ['amd64', 'arm64'])",
				},
				"push": map[string]interface{}{
					"type":        "boolean",
					"description": "Push to registry after build",
					"default":     false,
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Override target registry",
				},
				"tags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Additional image tags",
				},
				"variables": map[string]interface{}{
					"type":        "object",
					"description": "Template variables as key-value pairs",
				},
				"no_cache": map[string]interface{}{
					"type":        "boolean",
					"description": "Disable build cache",
					"default":     false,
				},
			},
			Required: []string{"template"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		template, ok := request.Params.Arguments["template"].(string)
		if !ok {
			return mcp.NewToolResultError("template must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		// Build options
		opts := client.BuildOptions{
			Template: template,
		}

		if target, ok := request.Params.Arguments["target"].(string); ok {
			opts.Target = target
		}

		if archsRaw, ok := request.Params.Arguments["architectures"].([]interface{}); ok {
			for _, arch := range archsRaw {
				if archStr, ok := arch.(string); ok {
					opts.Architectures = append(opts.Architectures, archStr)
				}
			}
		}

		if push, ok := request.Params.Arguments["push"].(bool); ok {
			opts.Push = push
		}

		if registry, ok := request.Params.Arguments["registry"].(string); ok {
			opts.Registry = registry
		}

		if tagsRaw, ok := request.Params.Arguments["tags"].([]interface{}); ok {
			for _, tag := range tagsRaw {
				if tagStr, ok := tag.(string); ok {
					opts.Tags = append(opts.Tags, tagStr)
				}
			}
		}

		if varsRaw, ok := request.Params.Arguments["variables"].(map[string]interface{}); ok {
			opts.Variables = make(map[string]string)
			for k, v := range varsRaw {
				opts.Variables[k] = fmt.Sprintf("%v", v)
			}
		}

		if noCache, ok := request.Params.Arguments["no_cache"].(bool); ok {
			opts.NoCache = noCache
		}

		output, err := wg.BuildTemplate(opts)
		if err != nil {
			logger.Errorf("Failed to build template: %v", err)
			// Clean up the error message a bit
			errMsg := strings.TrimSpace(output)
			return mcp.NewToolResultError(fmt.Sprintf("Build failed: %v\n\n%s", err, errMsg)), nil
		}

		logger.Infof("Template %s built successfully", template)
		return mcp.NewToolResultText(fmt.Sprintf("Build successful:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
