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

func createManifest(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "create_manifest",
		Description: "Create a multi-architecture image manifest for combining platform-specific images",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"image_name": map[string]interface{}{
					"type":        "string",
					"description": "Image name (e.g., 'my-app')",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Target registry (e.g., 'ghcr.io')",
				},
				"tags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Image tags (e.g., ['latest', 'v1.0'])",
				},
				"digest_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing digest files from builds",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Registry namespace (e.g., 'myorg')",
				},
				"required_architectures": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Required architectures (e.g., ['amd64', 'arm64'])",
				},
			},
			Required: []string{"image_name", "registry", "tags"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		imageName, ok := request.Params.Arguments["image_name"].(string)
		if !ok {
			return mcp.NewToolResultError("image_name must be a string"), nil
		}

		registry, ok := request.Params.Arguments["registry"].(string)
		if !ok {
			return mcp.NewToolResultError("registry must be a string"), nil
		}

		tagsRaw, ok := request.Params.Arguments["tags"].([]interface{})
		if !ok || len(tagsRaw) == 0 {
			return mcp.NewToolResultError("tags must be a non-empty array"), nil
		}

		var tags []string
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}

		if len(tags) == 0 {
			return mcp.NewToolResultError("tags must contain at least one string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		opts := client.ManifestOptions{
			ImageName: imageName,
			Registry:  registry,
			Tags:      tags,
		}

		if digestDir, ok := request.Params.Arguments["digest_dir"].(string); ok {
			opts.DigestDir = digestDir
		}

		if namespace, ok := request.Params.Arguments["namespace"].(string); ok {
			opts.Namespace = namespace
		}

		if archsRaw, ok := request.Params.Arguments["required_architectures"].([]interface{}); ok {
			for _, arch := range archsRaw {
				if archStr, ok := arch.(string); ok {
					opts.RequiredArchitectures = append(opts.RequiredArchitectures, archStr)
				}
			}
		}

		output, err := wg.CreateManifest(opts)
		if err != nil {
			logger.Errorf("Failed to create manifest: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Manifest creation failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Multi-arch manifest created successfully for %s", imageName)
		return mcp.NewToolResultText(fmt.Sprintf("Manifest created successfully:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
