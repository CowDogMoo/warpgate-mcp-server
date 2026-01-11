// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all available tools with the MCP server
func RegisterTools(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	// Warpgate CLI tools
	warpgateBuild(s, logger, warpgatePath)
	warpgateBuildStreaming(s, logger, warpgatePath)
	warpgateValidate(s, logger, warpgatePath)
	warpgateInit(s, logger, warpgatePath)

	// Template registry tools
	warpgateTemplatesList(s, logger, warpgatePath)
	warpgateTemplatesInfo(s, logger, warpgatePath)
	warpgateTemplatesAdd(s, logger, warpgatePath)
	warpgateTemplatesRemove(s, logger, warpgatePath)

	// Template creation
	createTemplate(s, logger, warpgatePath)

	// Manifest operations
	warpgateManifestsCreate(s, logger, warpgatePath)
	warpgateManifestsPush(s, logger, warpgatePath)

	// Config management
	warpgateConfigGet(s, logger, warpgatePath)
	warpgateConfigSet(s, logger, warpgatePath)
	warpgateConfigShow(s, logger, warpgatePath)

	// Conversion tools
	warpgateConvert(s, logger, warpgatePath)

	// Schema validation
	warpgateSchemaValidate(s, logger, warpgatePath)

	// Registry operations
	warpgateRegistryList(s, logger, warpgatePath)
	warpgateRegistryInspect(s, logger, warpgatePath)
	warpgateRegistryDelete(s, logger, warpgatePath)
	warpgateRegistryCopy(s, logger, warpgatePath)

	// Workflow tools
	listTasks(s, logger, warpgatePath)
	runTask(s, logger, warpgatePath)
	runImageBuilder(s, logger, warpgatePath)
}
