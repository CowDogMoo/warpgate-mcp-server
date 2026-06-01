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
				"verify_registry": map[string]interface{}{
					"type":        "boolean",
					"description": "Verify the digests exist in the registry before assembling the manifest. CLI default is true; set false to skip verification.",
				},
				"verify_concurrency": map[string]interface{}{
					"type":        "integer",
					"description": "Concurrent digest-verification requests. CLI default is 5; max 20.",
				},
				"max_age": map[string]interface{}{
					"type":        "string",
					"description": "Reject digest files older than this duration (e.g. '1h', '30m').",
				},
				"require_arch": map[string]interface{}{
					"type":        "array",
					"description": "Architectures that must be present in the assembled manifest (e.g. ['amd64', 'arm64']).",
					"items":       map[string]interface{}{"type": "string"},
				},
				"best_effort": map[string]interface{}{
					"type":        "boolean",
					"description": "Build the manifest with whatever architectures are available rather than failing when one is missing.",
				},
				"annotations": map[string]interface{}{
					"type":        "array",
					"description": "OCI annotations to apply, as 'key=value' strings.",
					"items":       map[string]interface{}{"type": "string"},
				},
				"labels": map[string]interface{}{
					"type":        "array",
					"description": "OCI labels to apply, as 'key=value' strings.",
					"items":       map[string]interface{}{"type": "string"},
				},
				"health_check": map[string]interface{}{
					"type":        "boolean",
					"description": "Perform a registry health check before assembling.",
				},
				"show_diff": map[string]interface{}{
					"type":        "boolean",
					"description": "If a manifest with the same name already exists, show a comparison/diff.",
				},
				"no_progress": map[string]interface{}{
					"type":        "boolean",
					"description": "Disable progress indicators.",
				},
				"quiet": map[string]interface{}{
					"type":        "boolean",
					"description": "Suppress informational output; only emit errors.",
				},
				"verbose": map[string]interface{}{
					"type":        "boolean",
					"description": "Verbose output with detailed progress.",
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
			Name:              name,
			Registry:          registry,
			Namespace:         request.GetString("namespace", ""),
			AuthFile:          request.GetString("auth_file", ""),
			Tags:              request.GetStringSlice("tags", nil),
			DigestDir:         request.GetString("digest_dir", ""),
			DryRun:            request.GetBool("dry_run", false),
			Force:             request.GetBool("force", false),
			VerifyConcurrency: request.GetInt("verify_concurrency", 0),
			MaxAge:            request.GetString("max_age", ""),
			RequireArch:       request.GetStringSlice("require_arch", nil),
			BestEffort:        request.GetBool("best_effort", false),
			Annotations:       request.GetStringSlice("annotations", nil),
			Labels:            request.GetStringSlice("labels", nil),
			HealthCheck:       request.GetBool("health_check", false),
			ShowDiff:          request.GetBool("show_diff", false),
			NoProgress:        request.GetBool("no_progress", false),
			Quiet:             request.GetBool("quiet", false),
			Verbose:           request.GetBool("verbose", false),
		}
		// verify_registry is tri-state (CLI default true). Only set if caller passed it.
		rawArgs := request.GetArguments()
		if raw, ok := rawArgs["verify_registry"]; ok {
			if b, ok := raw.(bool); ok {
				opts.VerifyRegistry = &b
			}
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

func warpgateManifestsList(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_manifests_list",
		Description: "List available manifest tags for an image in a container registry.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name to list tags for (e.g., 'attack-box').",
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

		opts := client.ManifestsListOptions{
			Name:      name,
			Registry:  registry,
			Namespace: request.GetString("namespace", ""),
			AuthFile:  request.GetString("auth_file", ""),
		}
		output, err := wg.WarpgateManifestsList(ctx, opts)
		if err != nil {
			logger.Errorf("Failed to list manifests: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list manifests: %v\n%s", err, output)), nil
		}
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}

func warpgateManifestsInspect(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_manifests_inspect",
		Description: "Inspect a multi-architecture manifest in a container registry. Returns architectures, digests, sizes, and annotations for each tag.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Image name to inspect (e.g., 'attack-box').",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Container registry, e.g. 'ghcr.io/cowdogmoo'. Required by the warpgate CLI.",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Image namespace/organization (optional).",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Tags to inspect. Defaults to ['latest'] when omitted.",
					"items":       map[string]interface{}{"type": "string"},
				},
				"auth_file": map[string]interface{}{
					"type":        "string",
					"description": "Path to a registry auth file (optional).",
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

		opts := client.ManifestsInspectOptions{
			Name:      name,
			Registry:  registry,
			Namespace: request.GetString("namespace", ""),
			Tags:      request.GetStringSlice("tags", nil),
			AuthFile:  request.GetString("auth_file", ""),
		}
		output, err := wg.WarpgateManifestsInspect(ctx, opts)
		if err != nil {
			logger.Errorf("Failed to inspect manifest: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to inspect manifest: %v\n%s", err, output)), nil
		}
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
