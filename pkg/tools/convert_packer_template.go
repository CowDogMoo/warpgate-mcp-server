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

func convertPackerTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "convert_packer_template",
		Description: "Convert a Packer template to warpgate.yaml format for migration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing Packer template files",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file path (default: warpgate.yaml)",
				},
				"author": map[string]interface{}{
					"type":        "string",
					"description": "Template author name",
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Template version (default: 1.0.0)",
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AMI target configuration",
					"default":     true,
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview conversion without writing files",
					"default":     false,
				},
			},
			Required: []string{"template_dir"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateDir, ok := request.Params.Arguments["template_dir"].(string)
		if !ok {
			return mcp.NewToolResultError("template_dir must be a string"), nil
		}

		wg, err := client.NewWarpgateClient(warpgatePath)
		if err != nil {
			logger.Errorf("Failed to create Warpgate client: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create Warpgate client: %v", err)), nil
		}

		opts := client.ConvertOptions{
			TemplateDir: templateDir,
			IncludeAMI:  true, // default
		}

		if output, ok := request.Params.Arguments["output"].(string); ok {
			opts.Output = output
		}

		if author, ok := request.Params.Arguments["author"].(string); ok {
			opts.Author = author
		}

		if version, ok := request.Params.Arguments["version"].(string); ok {
			opts.Version = version
		}

		if includeAMI, ok := request.Params.Arguments["include_ami"].(bool); ok {
			opts.IncludeAMI = includeAMI
		}

		if dryRun, ok := request.Params.Arguments["dry_run"].(bool); ok {
			opts.DryRun = dryRun
		}

		output, err := wg.ConvertPackerTemplate(opts)
		if err != nil {
			logger.Errorf("Failed to convert Packer template: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Conversion failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Packer template converted successfully")
		return mcp.NewToolResultText(fmt.Sprintf("Conversion successful:\n%s", output)), nil
	}

	s.AddTool(tool, handler)
}
