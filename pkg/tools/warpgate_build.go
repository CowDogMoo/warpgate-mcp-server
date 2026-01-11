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

func warpgateBuild(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_build",
		Description: "Build a container image or AMI using the warpgate CLI. Supports building from local config files, named templates from the registry, or git repositories.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Template name from registry, config file path, or warpgate.yaml location",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Build target type: 'container' or 'ami'",
					"enum":        []string{"container", "ami"},
				},
				"architectures": map[string]interface{}{
					"type":        "array",
					"description": "Target architectures to build (e.g., ['amd64', 'arm64'])",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"push": map[string]interface{}{
					"type":        "boolean",
					"description": "Push image to registry after build",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry to push to",
				},
				"vars": map[string]interface{}{
					"type":        "object",
					"description": "Variable overrides as key-value pairs",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Additional tags to apply to the image",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"no_cache": map[string]interface{}{
					"type":        "boolean",
					"description": "Disable build caching",
				},
				"save_digests": map[string]interface{}{
					"type":        "boolean",
					"description": "Save image digests to files after push",
				},
				"digest_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory to save digest files (default: current directory)",
				},
			},
			Required: []string{"template"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		template, ok := request.Params.Arguments["template"].(string)
		if !ok || template == "" {
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

		// Build options
		opts := client.BuildOptions{}

		if target, ok := request.Params.Arguments["target"].(string); ok {
			opts.Target = target
		}

		if archs, ok := request.Params.Arguments["architectures"].([]interface{}); ok {
			for _, arch := range archs {
				if a, ok := arch.(string); ok {
					opts.Architectures = append(opts.Architectures, a)
				}
			}
		}

		if push, ok := request.Params.Arguments["push"].(bool); ok {
			opts.Push = push
		}

		if registry, ok := request.Params.Arguments["registry"].(string); ok {
			opts.Registry = registry
		}

		if vars, ok := request.Params.Arguments["vars"].(map[string]interface{}); ok {
			opts.Vars = make(map[string]string)
			for k, v := range vars {
				if val, ok := v.(string); ok {
					opts.Vars[k] = val
				}
			}
		}

		if tags, ok := request.Params.Arguments["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if t, ok := tag.(string); ok {
					opts.Tags = append(opts.Tags, t)
				}
			}
		}

		if noCache, ok := request.Params.Arguments["no_cache"].(bool); ok {
			opts.NoCache = noCache
		}

		if saveDigests, ok := request.Params.Arguments["save_digests"].(bool); ok {
			opts.SaveDigests = saveDigests
		}

		if digestDir, ok := request.Params.Arguments["digest_dir"].(string); ok {
			opts.DigestDir = digestDir
		}

		output, err := wg.WarpgateBuild(template, opts)
		if err != nil {
			logger.Errorf("Build failed: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Build failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Build completed successfully for template: %s", template)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
