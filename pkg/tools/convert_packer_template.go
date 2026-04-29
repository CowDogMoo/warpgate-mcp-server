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

func convertPackerTemplate(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name:        "convert_packer_template",
		Description: "Convert an existing Packer HCL template to a warpgate.yaml. Use 'dry_run' to preview without writing.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing Packer .pkr.hcl files (variables, docker, ami)",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file path (default: <template_dir>/warpgate.yaml)",
				},
				"author": map[string]interface{}{
					"type":        "string",
					"description": "Template author",
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Template version (default from config)",
				},
				"base_image": map[string]interface{}{
					"type":        "string",
					"description": "Override base image (default: extracted from template)",
				},
				"license": map[string]interface{}{
					"type":        "string",
					"description": "Template license (default from config)",
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AMI target configuration",
					"default":     true,
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Print converted YAML without writing",
				},
			},
			Required: []string{"template_dir"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		dir := argString(args, "template_dir", "")
		if dir == "" {
			return mcp.NewToolResultError("template_dir is required"), nil
		}

		opts := client.ConvertOptions{
			TemplateDir: dir,
			Output:      argString(args, "output", ""),
			Author:      argString(args, "author", ""),
			Version:     argString(args, "version", ""),
			BaseImage:   argString(args, "base_image", ""),
			License:     argString(args, "license", ""),
			IncludeAMI:  argBoolPtr(args, "include_ami"),
			DryRun:      argBool(args, "dry_run", false),
		}

		out, err := wg.ConvertPackerTemplate(ctx, opts)
		if err != nil {
			logger.Errorf("convert_packer_template: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
