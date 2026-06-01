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

// buildToolProperties returns the InputSchema properties shared by warpgate_build
// and warpgate_build_streaming. They wrap the same CLI command, so their input
// surface must stay aligned.
func buildToolProperties() map[string]interface{} {
	return map[string]interface{}{
		"template": map[string]interface{}{
			"type":        "string",
			"description": "Template name from registry, config file path, or warpgate.yaml location",
		},
		"from_git": map[string]interface{}{
			"type":        "string",
			"description": "Load template from a git URL (alternative to --template).",
		},
		"target": map[string]interface{}{
			"type":        "string",
			"description": "Build target type. Default is read from the template.",
			"enum":        []string{"container", "ami", "azure", "proxmox"},
		},
		"architectures": map[string]interface{}{
			"type":        "array",
			"description": "Target architectures (e.g., ['amd64', 'arm64'])",
			"items":       map[string]interface{}{"type": "string"},
		},
		"push": map[string]interface{}{
			"type":        "boolean",
			"description": "Push image with tag after build. Mutually exclusive with push_digest.",
		},
		"push_digest": map[string]interface{}{
			"type":        "boolean",
			"description": "Push image by digest only, without creating a tag. Requires registry. Mutually exclusive with push.",
		},
		"registry": map[string]interface{}{
			"type":        "string",
			"description": "Container registry to push to.",
		},
		"vars": map[string]interface{}{
			"type":                 "object",
			"description":          "Template variable overrides as key-value pairs.",
			"additionalProperties": map[string]interface{}{"type": "string"},
		},
		"var_files": map[string]interface{}{
			"type":        "array",
			"description": "Load variables from one or more YAML files.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"build_args": map[string]interface{}{
			"type":        "array",
			"description": "Container build-args as 'key=value' strings.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"tags": map[string]interface{}{
			"type":        "array",
			"description": "Additional tags to apply to the image.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"no_cache": map[string]interface{}{
			"type":        "boolean",
			"description": "Disable build caching.",
		},
		"save_digests": map[string]interface{}{
			"type":        "boolean",
			"description": "Save image digests to files after push (for later use by warpgate_manifests_create).",
		},
		"digest_dir": map[string]interface{}{
			"type":        "string",
			"description": "Directory to save digest files (default: current directory).",
		},

		// AWS / AMI
		"region": map[string]interface{}{
			"type":        "string",
			"description": "AMI builds only: AWS region (overrides config).",
		},
		"instance_type": map[string]interface{}{
			"type":        "string",
			"description": "AMI builds only: EC2 instance type (overrides config).",
		},
		"force": map[string]interface{}{
			"type":        "boolean",
			"description": "AMI builds only: force recreation of existing AWS resources.",
		},
		"cleanup": map[string]interface{}{
			"type":        "boolean",
			"description": "Delete all build resources after a successful build (default: false for AMI, true for Azure).",
		},
		"dry_run": map[string]interface{}{
			"type":        "boolean",
			"description": "AMI builds only: validate configuration without creating resources.",
		},
		"regions": map[string]interface{}{
			"type":        "array",
			"description": "AMI builds only: build in multiple regions.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"parallel_regions": map[string]interface{}{
			"type":        "boolean",
			"description": "AMI builds only: build in all regions in parallel (default: sequential).",
		},
		"copy_to_regions": map[string]interface{}{
			"type":        "array",
			"description": "AMI builds only: copy the resulting AMI to additional regions after build.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"stream_logs": map[string]interface{}{
			"type":        "boolean",
			"description": "AMI builds only: stream CloudWatch/SSM logs from the build instance.",
		},
		"show_ec2_status": map[string]interface{}{
			"type":        "boolean",
			"description": "AMI builds only: show EC2 instance status during build.",
		},
		"output_manifest": map[string]interface{}{
			"type":        "string",
			"description": "Write a build-manifest JSON describing the result to this file.",
		},

		// Azure
		"azure_subscription": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: subscription ID.",
		},
		"azure_location": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: region for the build.",
		},
		"azure_resource_group": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: resource group.",
		},
		"azure_gallery": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: Compute Gallery name.",
		},
		"azure_image_definition": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: gallery image definition.",
		},
		"azure_vm_size": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: VM size used by Azure Image Builder.",
		},
		"azure_identity_id": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: user-assigned managed identity resource ID.",
		},
		"azure_target_regions": map[string]interface{}{
			"type":        "array",
			"description": "Azure builds only: regions to replicate the gallery image version to.",
			"items":       map[string]interface{}{"type": "string"},
		},
		"azure_subnet_id": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: pre-existing subnet resource ID for the build VM. Disables AIB's public IP.",
		},
		"azure_proxy_vm_size": map[string]interface{}{
			"type":        "string",
			"description": "Azure builds only: VM size for the AIB proxy when azure_subnet_id is set.",
		},

		// Proxmox
		"proxmox_endpoint": map[string]interface{}{
			"type":        "string",
			"description": "Proxmox builds only: API endpoint.",
		},
		"proxmox_node": map[string]interface{}{
			"type":        "string",
			"description": "Proxmox builds only: node name.",
		},
		"proxmox_storage": map[string]interface{}{
			"type":        "string",
			"description": "Proxmox builds only: storage for cloned disks.",
		},
		"proxmox_pool": map[string]interface{}{
			"type":        "string",
			"description": "Proxmox builds only: resource pool.",
		},
	}
}

// extractBuildOptions pulls a client.BuildOptions out of an MCP request. Shared
// by warpgate_build and warpgate_build_streaming.
func extractBuildOptions(request mcp.CallToolRequest) client.BuildOptions {
	opts := client.BuildOptions{
		FromGit:       request.GetString("from_git", ""),
		Target:        request.GetString("target", ""),
		Architectures: request.GetStringSlice("architectures", nil),
		Push:          request.GetBool("push", false),
		PushDigest:    request.GetBool("push_digest", false),
		Registry:      request.GetString("registry", ""),
		VarFiles:      request.GetStringSlice("var_files", nil),
		BuildArgs:     request.GetStringSlice("build_args", nil),
		Tags:          request.GetStringSlice("tags", nil),
		NoCache:       request.GetBool("no_cache", false),
		SaveDigests:   request.GetBool("save_digests", false),
		DigestDir:     request.GetString("digest_dir", ""),

		Region:          request.GetString("region", ""),
		InstanceType:    request.GetString("instance_type", ""),
		Force:           request.GetBool("force", false),
		Cleanup:         request.GetBool("cleanup", false),
		DryRun:          request.GetBool("dry_run", false),
		Regions:         request.GetStringSlice("regions", nil),
		ParallelRegions: request.GetBool("parallel_regions", false),
		CopyToRegions:   request.GetStringSlice("copy_to_regions", nil),
		StreamLogs:      request.GetBool("stream_logs", false),
		ShowEC2Status:   request.GetBool("show_ec2_status", false),
		OutputManifest:  request.GetString("output_manifest", ""),

		AzureSubscription: request.GetString("azure_subscription", ""),
		AzureLocation:     request.GetString("azure_location", ""),
		AzureResourceGrp:  request.GetString("azure_resource_group", ""),
		AzureGallery:      request.GetString("azure_gallery", ""),
		AzureImageDef:     request.GetString("azure_image_definition", ""),
		AzureVMSize:       request.GetString("azure_vm_size", ""),
		AzureIdentityID:   request.GetString("azure_identity_id", ""),
		AzureTargetRegion: request.GetStringSlice("azure_target_regions", nil),
		AzureSubnetID:     request.GetString("azure_subnet_id", ""),
		AzureProxyVMSize:  request.GetString("azure_proxy_vm_size", ""),

		ProxmoxEndpoint: request.GetString("proxmox_endpoint", ""),
		ProxmoxNode:     request.GetString("proxmox_node", ""),
		ProxmoxStorage:  request.GetString("proxmox_storage", ""),
		ProxmoxPool:     request.GetString("proxmox_pool", ""),
	}

	// Vars is an object<string,string>, not a primitive — pull it from raw args.
	args := request.GetArguments()
	if vars, ok := args["vars"].(map[string]interface{}); ok {
		opts.Vars = make(map[string]string, len(vars))
		for k, v := range vars {
			if val, ok := v.(string); ok {
				opts.Vars[k] = val
			}
		}
	}
	return opts
}

func warpgateBuild(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "warpgate_build",
		Description: "Build a container image, AMI, Azure Compute Gallery image, or Proxmox template using the warpgate CLI. Supports building from local config files, named templates from the registry, or git repositories.",
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
		output, err := wg.WarpgateBuild(ctx, template, opts)
		if err != nil {
			logger.Errorf("Build failed: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Build failed: %v\n%s", err, output)), nil
		}

		logger.Infof("Build completed successfully for template: %s", template)
		return mcp.NewToolResultText(output), nil
	}

	s.AddTool(tool, handler)
}
