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

func warpgateManifestsCreate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_manifests_create",
		Description: "Create and push a multi-architecture manifest. Discovers digest files produced by prior `warpgate build --save-digests` runs (named digest-<name>-<arch>.txt) under digest_dir, builds a multi-arch manifest list, and pushes it to the registry.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name to publish (e.g., 'attack-box').",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry, e.g. 'ghcr.io/cowdogmoo'. Required by the warpgate CLI.",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Image namespace/organization (optional).",
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Path to a registry auth file (optional).",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Tags to publish. Defaults to ['latest'] when omitted.",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"digest_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing digest files. Defaults to the current directory.",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview the manifest without pushing.",
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force recreation even if the manifest already exists in the registry.",
				},
			},
			Required: []string{"name", "registry"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required and must be a string"), nil
		}
		registry := request.GetString("registry", "")
		if registry == "" {
			return mcp.NewToolResultError("registry is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 3.0.0"), nil
		}

		opts := client.ManifestsCreateOptions{
			Name:      name,
			Registry:  registry,
			Namespace: request.GetString("namespace", ""),
			AuthFile:  request.GetString("auth_file", ""),
			Tags:      request.GetStringSlice("tags", nil),
			DigestDir: request.GetString("digest_dir", ""),
			DryRun:    request.GetBool("dry_run", false),
			Force:     request.GetBool("force", false),
		}

		output, err := wg.WarpgateManifestsCreate(ctx, opts)
		if err != nil {
			logger.Errorf("Failed to create manifest: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create manifest: %v\n%s", err, output)), nil
		}

		logger.Infof("Successfully created manifest: %s", name)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
