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

func warpgateConvert(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_convert",
		Description: "Convert a Packer template to warpgate YAML format. Helps migrate existing Packer infrastructure to the new warpgate configuration system.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Path to the Packer template file or directory to convert",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output path for the generated warpgate.yaml (default: <template-dir>/warpgate.yaml)",
				},
				"author": map[string]interface{}{
					"type":        "string",
					"description": "Template author to record in the generated metadata.",
				},
				"license": map[string]interface{}{
					"type":        "string",
					"description": "Template license (defaults to value from warpgate config).",
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Template version (defaults to value from warpgate config).",
				},
				"base_image": map[string]interface{}{
					"type":        "string",
					"description": "Override the base image instead of extracting it from the source Packer template.",
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AMI target configuration in the generated YAML. Defaults to true; set false to skip the AMI target.",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Print converted YAML to stdout without writing the output file.",
				},
			},
			Required: []string{"source"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := request.GetString("source", "")
		if source == "" {
			return mcp.NewToolResultError("source is required and must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		if !wg.IsCLIAvailable() {
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 3.0.0"), nil
		}

		opts := client.ConvertOptions{
			Source:    source,
			Output:    request.GetString("output", ""),
			Author:    request.GetString("author", ""),
			License:   request.GetString("license", ""),
			Version:   request.GetString("version", ""),
			BaseImage: request.GetString("base_image", ""),
			DryRun:    request.GetBool("dry_run", false),
		}
		// include_ami is tri-state: unset (CLI default true), explicitly true, explicitly false.
		// Only override if the caller passed the field.
		args := request.GetArguments()
		if raw, ok := args["include_ami"]; ok {
			if b, ok := raw.(bool); ok {
				opts.IncludeAMI = &b
			}
		}

		result, err := wg.WarpgateConvert(ctx, opts)
		if err != nil {
			logger.Errorf("Failed to convert template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert template: %v\n%s", err, result)), nil
		}

		logger.Infof("Successfully converted template: %s", source)
		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}
