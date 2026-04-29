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

func cleanupResources(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name: "cleanup_aws_resources",
		Description: "Clean up AWS Image Builder resources created by warpgate. " +
			"Use 'dry_run' first to preview. Always non-interactive (--yes is forced) — only run when explicitly requested.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"build_name": map[string]interface{}{
					"type":        "string",
					"description": "Build name to clean up (e.g. 'attack-box'). Omit when 'all' is true.",
				},
				"all": map[string]interface{}{
					"type":        "boolean",
					"description": "Operate on all warpgate-created resources",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview without deleting (strongly recommended first)",
				},
				"region": map[string]interface{}{
					"type":        "string",
					"description": "AWS region (defaults to config)",
				},
				"versions": map[string]interface{}{
					"type":        "boolean",
					"description": "Clean up old component versions instead of all resources",
				},
				"keep": map[string]interface{}{
					"type":        "integer",
					"description": "Number of component versions to keep when versions=true (default 3)",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		opts := client.CleanupOptions{
			BuildName: argString(args, "build_name", ""),
			All:       argBool(args, "all", false),
			DryRun:    argBool(args, "dry_run", false),
			Region:    argString(args, "region", ""),
			Versions:  argBool(args, "versions", false),
			Keep:      argInt(args, "keep", 0),
			Yes:       true, // MCP context is non-interactive
		}
		if !opts.All && opts.BuildName == "" {
			return mcp.NewToolResultError("either build_name or all=true is required"), nil
		}

		out, err := wg.Cleanup(ctx, opts)
		if err != nil {
			logger.Errorf("cleanup_aws_resources: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
