// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all available tools with the MCP server
func RegisterTools(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	// Template management tools
	listTemplates(s, logger, warpgatePath)
	getTemplateInfo(s, logger, warpgatePath)
	initTemplate(s, logger, warpgatePath)
	validateTemplate(s, logger, warpgatePath)
	buildTemplate(s, logger, warpgatePath)

	// Template discovery and source management
	searchTemplates(s, logger, warpgatePath)
	addTemplateSource(s, logger, warpgatePath)
	removeTemplateSource(s, logger, warpgatePath)
	updateTemplateCache(s, logger, warpgatePath)

	// Build and conversion tools
	convertPackerTemplate(s, logger, warpgatePath)
	createManifest(s, logger, warpgatePath)

	// Workflow tools (kept for backward compatibility)
	listTasks(s, logger, warpgatePath)
	runTask(s, logger, warpgatePath)
	runPreCommit(s, logger, warpgatePath)
	runImageBuilder(s, logger, warpgatePath)
}
