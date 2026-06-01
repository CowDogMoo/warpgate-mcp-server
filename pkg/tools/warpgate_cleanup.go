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

func warpgateCleanup(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_cleanup",
		Description: "DESTRUCTIVE: Delete AWS resources created by prior `warpgate build` runs (AMIs, snapshots, EC2 components). Always invoke with dry_run=true first to preview what would be deleted. Pass yes=true to skip the CLI's interactive confirmation; without it the call will hang waiting for input.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"build_name": map[string]interface{}{
					"type":        "string",
					"description": "Scope cleanup to resources tagged with this build name. Omit (with all=true) to clean up across all warpgate builds.",
				},
				"region": map[string]interface{}{
					"type":        "string",
					"description": "AWS region (defaults to value from warpgate config).",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview what would be deleted without actually deleting. Strongly recommended on the first call.",
				},
				"all": map[string]interface{}{
					"type":        "boolean",
					"description": "Operate on every warpgate-tagged resource in the account, not just one build.",
				},
				"versions": map[string]interface{}{
					"type":        "boolean",
					"description": "Clean up old component versions (keeping the most recent N from keep) instead of deleting every resource.",
				},
				"keep": map[string]interface{}{
					"type":        "integer",
					"description": "When versions=true, how many recent versions to retain (CLI default: 3).",
				},
				"yes": map[string]interface{}{
					"type":        "boolean",
					"description": "Skip confirmation prompts. Required for non-interactive callers (including MCP).",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}
		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 3.0.0"), nil
		}

		opts := client.CleanupOptions{
			BuildName:    request.GetString("build_name", ""),
			Region:       request.GetString("region", ""),
			DryRun:       request.GetBool("dry_run", false),
			All:          request.GetBool("all", false),
			Versions:     request.GetBool("versions", false),
			KeepVersions: request.GetInt("keep", 0),
			Yes:          request.GetBool("yes", false),
		}

		output, err := wg.WarpgateCleanup(ctx, opts)
		if err != nil {
			logger.Errorf("Cleanup failed: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Cleanup failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Cleanup completed (build_name=%q dry_run=%t)", opts.BuildName, opts.DryRun)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
