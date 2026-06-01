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
		Description: "Build a container image, AMI, Azure Compute Gallery image, or Proxmox template with real-time progress output. Uses MCP logging notifications to stream build output line by line. Ideal for long-running builds where progress visibility is important.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: buildToolProperties(),
			Required:   []string{"template"},
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
			return mcp.NewToolResultError("warpgate CLI is not available. Please install warpgate >= 3.0.0"), nil
		}

		opts := extractBuildOptions(request)

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

		// Create callback for streaming output
		callback := func(line string) {
			lineCount++

			// Log to our logger
			logger.Infof("[BUILD] %s", line)

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
		output, err := wg.WarpgateBuildStreaming(ctx, template, opts, callback)

		// Write complete output to log file (EXACTLY what you'd see in terminal)
		if buildLogFile != nil {
			_, _ = fmt.Fprintf(buildLogFile, "%s", output)
			_ = buildLogFile.Sync()
		}
		if err != nil {
			logger.Errorf("Build failed: %v", err)

			// Write failure marker to log file
			if buildLogFile != nil {
				_, _ = fmt.Fprintf(buildLogFile, "\n--- BUILD FAILED ---\nError: %v\n", err)
				_ = buildLogFile.Sync()
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
			_ = buildLogFile.Sync()
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
