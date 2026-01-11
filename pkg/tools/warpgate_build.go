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

		// Build options
		opts := client.BuildOptions{
			Target:        request.GetString("target", ""),
			Architectures: request.GetStringSlice("architectures", nil),
			Push:          request.GetBool("push", false),
			Registry:      request.GetString("registry", ""),
			Tags:          request.GetStringSlice("tags", nil),
			NoCache:       request.GetBool("no_cache", false),
			SaveDigests:   request.GetBool("save_digests", false),
			DigestDir:     request.GetString("digest_dir", ""),
		}

		// Handle vars map separately since it needs conversion
		args := request.GetArguments()
		if vars, ok := args["vars"].(map[string]interface{}); ok {
			opts.Vars = make(map[string]string)
			for k, v := range vars {
				if val, ok := v.(string); ok {
					opts.Vars[k] = val
				}
			}
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
