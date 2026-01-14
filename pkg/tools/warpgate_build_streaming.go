// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func warpgateBuildStreaming(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_build_streaming",
		Description: "Build a container image or AMI with real-time progress output. Uses MCP logging notifications to stream build output line by line. Ideal for long-running builds where progress visibility is important.",
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		// Get the MCP server from context for sending notifications
		mcpServer := server.ServerFromContext(ctx)

		// Create build-specific log file
		buildLogFile, buildLogPath, err := logging.CreateBuildLogFile(template)
		if err != nil {
			logger.Warnf("Failed to create build log file: %v (continuing without file logging)", err)
		}
		defer func() {
			if buildLogFile != nil {
				_ = buildLogFile.Close()
			}
		}()

		// Track line count for progress
		var lineCount int
		var outputLines []string

		// Create callback for streaming output
		callback := func(line string) {
			lineCount++
			outputLines = append(outputLines, line)

			// Log to our logger
			logger.Infof("[BUILD] %s", line)

			// Write to build-specific log file
			if buildLogFile != nil {
				_, _ = fmt.Fprintf(buildLogFile, "%s\n", line)
			}

			// Send logging notification to MCP client if server is available
			if mcpServer != nil {
				// Determine log level based on line content
				lineLower := strings.ToLower(line)
				var level string
				switch {
				case strings.Contains(lineLower, "error") || strings.Contains(lineLower, "failed"):
					level = "error"
				case strings.Contains(lineLower, "warning") || strings.Contains(lineLower, "warn"):
					level = "warning"
				case strings.Contains(lineLower, "step") || strings.Contains(lineLower, "building"):
					level = "notice"
				default:
					level = "info"
				}

				// Send notification (best effort, ignore errors)
				_ = mcpServer.SendNotificationToClient(ctx, "notifications/logging/message", map[string]interface{}{
					"level":  level,
					"logger": "warpgate.build",
					"data":   line,
				})
			}
		}

		// Run the streaming build
		output, err := wg.WarpgateBuildStreaming(template, opts, callback)
		if err != nil {
			logger.Errorf("Build failed: %v", err)

			// Write error to log file
			if buildLogFile != nil {
				_, _ = fmt.Fprintf(buildLogFile, "\n--- BUILD FAILED ---\nError: %v\n", err)
			}

			// Send error notification
			if mcpServer != nil {
				_ = mcpServer.SendNotificationToClient(ctx, "notifications/logging/message", map[string]interface{}{
					"level":  "error",
					"logger": "warpgate.build",
					"data":   fmt.Sprintf("Build failed: %v", err),
				})
			}

			errorMsg := fmt.Sprintf("Build failed: %v\n%s", err, output)
			if buildLogPath != "" {
				errorMsg += fmt.Sprintf("\n\nFull build log: %s", buildLogPath)
			}
			return mcp.NewToolResultError(errorMsg), nil
		}

		// Write success footer to log file
		if buildLogFile != nil {
			_, _ = fmt.Fprintf(buildLogFile, "\n--- BUILD COMPLETED SUCCESSFULLY ---\n")
		}

		// Send completion notification
		if mcpServer != nil {
			_ = mcpServer.SendNotificationToClient(ctx, "notifications/logging/message", map[string]interface{}{
				"level":  "notice",
				"logger": "warpgate.build",
				"data":   fmt.Sprintf("Build completed successfully (%d lines of output)", lineCount),
			})
		}

		logger.Infof("Build completed successfully for template: %s (%d lines)", template, lineCount)

		// Include log file path in response
		successMsg := output
		if buildLogPath != "" {
			successMsg += fmt.Sprintf("\n\nFull build log saved to: %s", buildLogPath)
		}
		return mcp.NewToolResultText(successMsg), nil
	}

	s.AddTool(tool, handler)
}
