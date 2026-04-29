// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func createManifest(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name: "create_manifest",
		Description: "Create and push a multi-architecture image manifest from per-arch digest files. " +
			"Digest files must follow the naming convention 'digest-{IMAGE_NAME}-{ARCH}.txt' in 'digest_dir'.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"image_name": map[string]interface{}{
					"type":        "string",
					"description": "Image name (e.g. 'attack-box')",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry (required, e.g. 'ghcr.io/cowdogmoo')",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Image namespace/organization",
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Path to registry auth file",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Image tags (default: ['latest'])",
				},
				"digest_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing digest files (default '.')",
				},
				"required_architectures": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Required architectures (e.g. ['amd64','arm64'])",
				},
				"best_effort": map[string]interface{}{
					"type":        "boolean",
					"description": "Create manifest with whatever architectures are available",
				},
				"annotations": map[string]interface{}{
					"type":                 "object",
					"description":          "OCI annotations as key/value map",
					"additionalProperties": true,
				},
				"labels": map[string]interface{}{
					"type":                 "object",
					"description":          "OCI labels as key/value map",
					"additionalProperties": true,
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview manifest without pushing",
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force recreation if manifest already exists",
				},
				"health_check": map[string]interface{}{
					"type":        "boolean",
					"description": "Perform a registry health check first",
				},
				"max_age": map[string]interface{}{
					"type":        "string",
					"description": "Maximum age of digest files (e.g. '1h', '30m')",
				},
				"show_diff": map[string]interface{}{
					"type":        "boolean",
					"description": "Show manifest diff if it already exists",
				},
				"verify_registry": map[string]interface{}{
					"type":        "boolean",
					"description": "Verify digests exist in registry (default true)",
				},
				"verify_concurrency": map[string]interface{}{
					"type":        "integer",
					"description": "Number of concurrent digest verifications (default 5, max 20)",
				},
			},
			Required: []string{"image_name", "registry"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		name := argString(args, "image_name", "")
		registry := argString(args, "registry", "")
		if name == "" {
			return mcp.NewToolResultError("image_name is required"), nil
		}
		if registry == "" {
			return mcp.NewToolResultError("registry is required"), nil
		}

		opts := client.ManifestOptions{
			ImageName:             name,
			Registry:              registry,
			Namespace:             argString(args, "namespace", ""),
			AuthFile:              argString(args, "auth_file", ""),
			Tags:                  argStringSlice(args, "tags"),
			DigestDir:             argString(args, "digest_dir", ""),
			RequiredArchitectures: argStringSlice(args, "required_architectures"),
			BestEffort:            argBool(args, "best_effort", false),
			Annotations:           argStringMap(args, "annotations"),
			Labels:                argStringMap(args, "labels"),
			DryRun:                argBool(args, "dry_run", false),
			Force:                 argBool(args, "force", false),
			HealthCheck:           argBool(args, "health_check", false),
			MaxAge:                argString(args, "max_age", ""),
			ShowDiff:              argBool(args, "show_diff", false),
			VerifyRegistry:        argBoolPtr(args, "verify_registry"),
			VerifyConcurrency:     argInt(args, "verify_concurrency", 0),
		}

		out, err := wg.CreateManifest(ctx, opts)
		if err != nil {
			logger.Errorf("create_manifest: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
