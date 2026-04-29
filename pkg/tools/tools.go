// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"fmt"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all warpgate tools with the MCP server.
func RegisterTools(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	// Templates
	listTemplates(s, logger, wg)
	searchTemplates(s, logger, wg)
	getTemplateInfo(s, logger, wg)
	initTemplate(s, logger, wg)
	validateTemplate(s, logger, wg)
	addTemplateSource(s, logger, wg)
	removeTemplateSource(s, logger, wg)
	updateTemplateCache(s, logger, wg)

	// Build & related
	buildTemplate(s, logger, wg)
	convertPackerTemplate(s, logger, wg)

	// Manifests (multi-arch)
	createManifest(s, logger, wg)
	inspectManifest(s, logger, wg)
	listManifests(s, logger, wg)

	// AWS cleanup
	cleanupResources(s, logger, wg)

	// Config
	showConfig(s, logger, wg)
	getConfig(s, logger, wg)
	setConfig(s, logger, wg)
}

// argString safely extracts a string argument with optional default.
func argString(args map[string]interface{}, key, fallback string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

// argBool safely extracts a bool argument.
func argBool(args map[string]interface{}, key string, fallback bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return fallback
}

// argBoolPtr returns &v if the key was explicitly provided, nil otherwise.
// Use this for tri-state flags that need to distinguish "unset" from "false".
func argBoolPtr(args map[string]interface{}, key string) *bool {
	if v, ok := args[key].(bool); ok {
		return &v
	}
	return nil
}

// argInt safely extracts an int argument.
func argInt(args map[string]interface{}, key string, fallback int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return fallback
}

// argStringSlice extracts an array of strings.
func argStringSlice(args map[string]interface{}, key string) []string {
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// argStringMap extracts a {string: string} map (coerces other scalars to string).
func argStringMap(args map[string]interface{}, key string) map[string]string {
	raw, ok := args[key].(map[string]interface{})
	if !ok {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
		} else {
			out[k] = fmt.Sprint(v)
		}
	}
	return out
}
