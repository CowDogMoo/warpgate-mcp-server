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

func buildTemplate(s *server.MCPServer, logger *logging.Logger, wg *client.WarpgateClient) {
	tool := mcp.Tool{
		Name: "build_template",
		Description: "Build a container image or AWS AMI from a warpgate template. " +
			"Provide one of: 'template' (named template from registry), 'config' (path to warpgate.yaml), or 'from_git' (git URL).",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Named template from registry (e.g. 'attack-box' or 'attack-box@v1.2.0')",
				},
				"config": map[string]interface{}{
					"type":        "string",
					"description": "Path to a local warpgate.yaml",
				},
				"from_git": map[string]interface{}{
					"type":        "string",
					"description": "Git URL to load a template from (supports //subdir suffix)",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Override target type",
					"enum":        []string{"container", "ami"},
				},
				"architectures": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Architectures to build (e.g. ['amd64','arm64'])",
				},
				"push": map[string]interface{}{
					"type":        "boolean",
					"description": "Push image to registry with tag after build",
				},
				"push_digest": map[string]interface{}{
					"type":        "boolean",
					"description": "Push by digest only (no tag); requires registry. Mutually exclusive with push.",
				},
				"registry": map[string]interface{}{
					"type":        "string",
					"description": "Override target registry (e.g. ghcr.io/myorg)",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Additional image tags",
				},
				"variables": map[string]interface{}{
					"type":                 "object",
					"description":          "Template variables as a key/value map (becomes --var KEY=VAL)",
					"additionalProperties": true,
				},
				"var_files": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Paths to YAML files containing template variables",
				},
				"build_args": map[string]interface{}{
					"type":                 "object",
					"description":          "Docker build args as key/value map",
					"additionalProperties": true,
				},
				"labels": map[string]interface{}{
					"type":                 "object",
					"description":          "OCI image labels as key/value map",
					"additionalProperties": true,
				},
				"no_cache": map[string]interface{}{
					"type":        "boolean",
					"description": "Disable all caching",
				},
				"cache_from": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "BuildKit cache sources (e.g. type=registry,ref=user/app:cache)",
				},
				"cache_to": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "BuildKit cache destinations",
				},
				"output_manifest": map[string]interface{}{
					"type":        "string",
					"description": "Write a build manifest JSON to this path",
				},
				"save_digests": map[string]interface{}{
					"type":        "boolean",
					"description": "Save image digests to files after push (for later 'create_manifest')",
				},
				"digest_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory to save digest files (default '.')",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Validate config without creating resources (AMI only)",
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force recreation of existing AWS resources (AMI only)",
				},
				"region": map[string]interface{}{
					"type":        "string",
					"description": "AWS region for AMI builds",
				},
				"regions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Build AMI in multiple regions",
				},
				"copy_to_regions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Copy AMI to additional regions after build",
				},
				"parallel_regions": map[string]interface{}{
					"type":        "boolean",
					"description": "Build all regions in parallel (default sequential)",
				},
				"instance_type": map[string]interface{}{
					"type":        "string",
					"description": "EC2 instance type for AMI builds",
				},
				"stream_logs": map[string]interface{}{
					"type":        "boolean",
					"description": "Stream CloudWatch/SSM logs from build instance (AMI only)",
				},
				"show_ec2_status": map[string]interface{}{
					"type":        "boolean",
					"description": "Show EC2 instance status during build (AMI only)",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments

		opts := client.BuildOptions{
			Template:        argString(args, "template", ""),
			Config:          argString(args, "config", ""),
			FromGit:         argString(args, "from_git", ""),
			Target:          argString(args, "target", ""),
			Architectures:   argStringSlice(args, "architectures"),
			Push:            argBool(args, "push", false),
			PushDigest:      argBool(args, "push_digest", false),
			Registry:        argString(args, "registry", ""),
			Tags:            argStringSlice(args, "tags"),
			Variables:       argStringMap(args, "variables"),
			VarFiles:        argStringSlice(args, "var_files"),
			BuildArgs:       argStringMap(args, "build_args"),
			Labels:          argStringMap(args, "labels"),
			NoCache:         argBool(args, "no_cache", false),
			CacheFrom:       argStringSlice(args, "cache_from"),
			CacheTo:         argStringSlice(args, "cache_to"),
			OutputManifest:  argString(args, "output_manifest", ""),
			SaveDigests:     argBool(args, "save_digests", false),
			DigestDir:       argString(args, "digest_dir", ""),
			DryRun:          argBool(args, "dry_run", false),
			Force:           argBool(args, "force", false),
			Region:          argString(args, "region", ""),
			Regions:         argStringSlice(args, "regions"),
			CopyToRegions:   argStringSlice(args, "copy_to_regions"),
			ParallelRegions: argBool(args, "parallel_regions", false),
			InstanceType:    argString(args, "instance_type", ""),
			StreamLogs:      argBool(args, "stream_logs", false),
			ShowEC2Status:   argBool(args, "show_ec2_status", false),
		}

		out, err := wg.BuildTemplate(ctx, opts)
		if err != nil {
			logger.Errorf("build_template: %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	s.AddTool(tool, handler)
}
