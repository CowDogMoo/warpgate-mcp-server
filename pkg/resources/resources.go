// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/client"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	configResourceURI   = "warpgate://config"
	schemaResourceURI   = "warpgate://schema/template"
	exampleResourceURI  = "warpgate://examples/template"
	templateSchemaURL   = "https://raw.githubusercontent.com/cowdogmoo/warpgate/main/schema/warpgate-template.json"
)

//go:embed embedded_example.yaml
var embeddedExample string

// RegisterResources wires resources to the MCP server.
func RegisterResources(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	configResource(s, logger, wg)
	schemaResource(s, logger)
	exampleResource(s, logger)
}

// configResource exposes the resolved warpgate configuration.
// Uses `warpgate config show` so callers see the same view warpgate uses
// (defaults + file + env + flags), not a hand-parsed yaml file.
func configResource(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	resource := mcp.Resource{
		URI:         configResourceURI,
		Name:        "Warpgate Configuration",
		Description: "Resolved warpgate configuration (defaults + config file + env + flags)",
		MIMEType:    "application/yaml",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		out, err := wg.ConfigShow(ctx)
		if err != nil {
			logger.Errorf("read config resource: %v", err)
			return nil, fmt.Errorf("read warpgate config: %w", err)
		}
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{URI: configResourceURI, MIMEType: "application/yaml"},
				Text:             out,
			},
		}, nil
	}
	s.AddResource(resource, handler)
}

// schemaResource fetches the warpgate.yaml JSON schema from upstream.
func schemaResource(s *server.MCPServer, logger *logging.Logger) {
	resource := mcp.Resource{
		URI:         schemaResourceURI,
		Name:        "Warpgate Template Schema",
		Description: "JSON schema for warpgate.yaml template files",
		MIMEType:    "application/json",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, templateSchemaURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build schema request: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Errorf("fetch schema: %v", err)
			return nil, fmt.Errorf("fetch schema: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch schema: HTTP %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read schema body: %w", err)
		}
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{URI: schemaResourceURI, MIMEType: "application/json"},
				Text:             string(body),
			},
		}, nil
	}
	s.AddResource(resource, handler)
}

// exampleResource serves an example warpgate.yaml. Prefers the on-disk example
// in the repo, falls back to an embedded copy so the resource is always available.
func exampleResource(s *server.MCPServer, logger *logging.Logger) {
	resource := mcp.Resource{
		URI:         exampleResourceURI,
		Name:        "Example Warpgate Template",
		Description: "An example warpgate.yaml demonstrating the template format",
		MIMEType:    "application/yaml",
	}

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		text := embeddedExample
		if cwd, err := os.Getwd(); err == nil {
			candidate := filepath.Join(cwd, "examples", "warpgate.yaml")
			if data, err := os.ReadFile(candidate); err == nil {
				text = string(data)
			}
		}
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{URI: exampleResourceURI, MIMEType: "application/yaml"},
				Text:             text,
			},
		}, nil
	}
	s.AddResource(resource, handler)
}
