// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

// Package prompts provides MCP prompt handlers for the warpgate-mcp-server.
//
// Prompts expose reusable, parameterized workflow recipes that any MCP client
// (Claude Code, Cursor, Continue, ChatGPT desktop, etc.) can list and invoke.
// They reference the tools and resources registered by this server and tell
// the calling agent how to chain them together for common operator workflows.
package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterPrompts registers all available prompts with the MCP server.
func RegisterPrompts(s *server.MCPServer, logger *logging.Logger) {
	bootstrapNewTemplate(s, logger)
	debugFailedBuild(s, logger)
	addProvisioner(s, logger)
	convertFromPacker(s, logger)
	publishMultiarchImage(s, logger)
}

func userMessage(text string) []mcp.PromptMessage {
	return []mcp.PromptMessage{
		mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(text)),
	}
}

func requireArg(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", fmt.Errorf("argument %q is required", key)
	}
	return v, nil
}

// renderBootstrapNewTemplate produces the prompt body. Pure function so the
// rendered output can be unit-tested without exercising the MCP server.
func renderBootstrapNewTemplate(name, from, output string) string {
	var fromLine, initArgs string
	if from != "" {
		fromLine = fmt.Sprintf("Fork from existing template %q as the starting point.\n", from)
		initArgs = fmt.Sprintf("name=%q, from=%q", name, from)
	} else {
		initArgs = fmt.Sprintf("name=%q", name)
	}
	if output != "" {
		initArgs += fmt.Sprintf(", output=%q", output)
	}

	return fmt.Sprintf(`Bootstrap a new warpgate template named %q.

%sFollow these steps using the warpgate MCP tools and resources:

1. Scaffold: call the warpgate_init tool with %s. This creates warpgate.yaml,
   README.md, and a scripts/ directory.
2. Read the scaffold: fetch the warpgate://template/config and warpgate://schema
   resources so you know which fields exist and which are required.
3. Configure warpgate.yaml: set base image, tags, provisioners, and any
   build-time variables the template needs. Confirm non-obvious choices with
   the user before writing them.
4. Add provisioning: drop scripts into scripts/ and reference them from the
   provisioners list in warpgate.yaml. Use the add_provisioner prompt for
   non-trivial additions.
5. Validate: call warpgate_validate against the new template directory. Fix
   any reported issues before continuing.
6. Schema check: call warpgate_schema_validate for a stricter pass.
7. First build: call warpgate_build_streaming so the user can watch progress.
   If the build fails, switch to the debug_failed_build prompt with the
   template path and the error summary.

Stop and ask the user before running steps that produce artifacts (build,
push) if there's any ambiguity in the scaffolded config.`, name, fromLine, initArgs)
}

func bootstrapNewTemplate(s *server.MCPServer, logger *logging.Logger) {
	prompt := mcp.NewPrompt(
		"bootstrap_new_template",
		mcp.WithPromptDescription(
			"Scaffold a new warpgate template, configure it, validate it, and run the first build.",
		),
		mcp.WithArgument("name",
			mcp.ArgumentDescription("Name of the new template (used as the directory name)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("from",
			mcp.ArgumentDescription("Optional existing template name to fork as a starting point"),
		),
		mcp.WithArgument("output",
			mcp.ArgumentDescription("Optional output directory for the scaffold (default: current directory)"),
		),
	)

	handler := func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name, err := requireArg(request.Params.Arguments, "name")
		if err != nil {
			logger.Errorf("bootstrap_new_template: %v", err)
			return nil, err
		}
		from := strings.TrimSpace(request.Params.Arguments["from"])
		output := strings.TrimSpace(request.Params.Arguments["output"])

		text := renderBootstrapNewTemplate(name, from, output)
		return mcp.NewGetPromptResult("Bootstrap a new warpgate template", userMessage(text)), nil
	}

	s.AddPrompt(prompt, handler)
}

func renderDebugFailedBuild(path, summary string) string {
	var summaryLine string
	if summary != "" {
		summaryLine = fmt.Sprintf("Reported failure: %s\n\n", summary)
	}

	return fmt.Sprintf(`Debug a failing warpgate build for the template at %s.

%sWork through these layers in order, stopping as soon as you find a concrete
fix to recommend:

1. Syntax & structure: call warpgate_validate against %s. This catches YAML
   parse errors, missing required fields, and obvious shape mistakes.
2. Schema conformance: call warpgate_schema_validate. Cross-reference any
   complaints against the warpgate://schema resource.
3. CLI version: fetch warpgate://cli/info and confirm the installed warpgate
   binary is recent enough for the features the template uses.
4. Effective config: fetch warpgate://config to see the merged config the CLI
   would actually use, including registry credentials and BuildKit settings.
5. Provisioners: read warpgate.yaml and the contents of scripts/. Common
   failure modes:
   - exit 127 / "command not found": missing package install step earlier in
     the provisioner chain
   - permission denied: script needs chmod +x, or needs to run as root
   - network errors: base image lacks DNS, or registry auth missing
   - architecture mismatch: building for an arch the base image doesn't
     publish
6. Reproduce minimally: if step 1-5 are clean, run warpgate_build_streaming
   so you can read the live output and pinpoint the failing layer.

Report findings as: (a) root cause, (b) the smallest fix, (c) the tool call
that would verify the fix. Do not modify the template until the user
approves the proposed change.`, path, summaryLine, path)
}

func debugFailedBuild(s *server.MCPServer, logger *logging.Logger) {
	prompt := mcp.NewPrompt(
		"debug_failed_build",
		mcp.WithPromptDescription(
			"Diagnose a failing warpgate build by walking through validation, schema, config, and common provisioner issues.",
		),
		mcp.WithArgument("template_path",
			mcp.ArgumentDescription("Path to the template directory containing warpgate.yaml"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("error_summary",
			mcp.ArgumentDescription("One-line summary of the failure, if known (e.g. 'provisioner exited 127')"),
		),
	)

	handler := func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		path, err := requireArg(request.Params.Arguments, "template_path")
		if err != nil {
			logger.Errorf("debug_failed_build: %v", err)
			return nil, err
		}
		summary := strings.TrimSpace(request.Params.Arguments["error_summary"])

		text := renderDebugFailedBuild(path, summary)
		return mcp.NewGetPromptResult("Debug a failed warpgate build", userMessage(text)), nil
	}

	s.AddPrompt(prompt, handler)
}

func renderAddProvisioner(path, ptype, desc string) string {
	var descLine string
	if desc != "" {
		descLine = fmt.Sprintf("Goal: %s\n\n", desc)
	}

	return fmt.Sprintf(`Add a %s provisioner to the warpgate template at %s.

%sFollow these steps:

1. Schema check: fetch warpgate://schema and find the provisioner shape for
   type %q. Note required vs. optional fields.
2. Read existing config: open %s/warpgate.yaml. Identify where the new
   provisioner should slot into the provisioners list (order matters —
   provisioners run sequentially and earlier ones can install prerequisites
   for later ones).
3. Create supporting files:
   - shell: drop a script into %s/scripts/ and reference it by path. Make it
     idempotent and use 'set -euo pipefail' at the top.
   - ansible: drop a playbook into %s/scripts/ (or a dedicated ansible/
     subdir) and reference it. Pin collection versions if relevant.
   - file: place the source file under the template directory and reference
     it with a relative path; specify the destination path on the image.
4. Edit warpgate.yaml: append the new provisioner entry. Preserve indentation
   and existing key order.
5. Validate: call warpgate_validate, then warpgate_schema_validate. Fix any
   complaints before continuing.
6. Test build (optional): if the user wants to verify, call
   warpgate_build_streaming and watch the new step run.

Show the user the proposed warpgate.yaml diff and any new files before
writing them.`, ptype, path, descLine, ptype, path, path, path)
}

func addProvisioner(s *server.MCPServer, logger *logging.Logger) {
	prompt := mcp.NewPrompt(
		"add_provisioner",
		mcp.WithPromptDescription(
			"Add a provisioner step (shell, ansible, or file) to an existing warpgate template.",
		),
		mcp.WithArgument("template_path",
			mcp.ArgumentDescription("Path to the template directory containing warpgate.yaml"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("provisioner_type",
			mcp.ArgumentDescription("Provisioner kind: shell, ansible, or file"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("description",
			mcp.ArgumentDescription("What the provisioner should do (e.g. 'install nmap and masscan')"),
		),
	)

	handler := func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		path, err := requireArg(request.Params.Arguments, "template_path")
		if err != nil {
			logger.Errorf("add_provisioner: %v", err)
			return nil, err
		}
		ptype, err := requireArg(request.Params.Arguments, "provisioner_type")
		if err != nil {
			logger.Errorf("add_provisioner: %v", err)
			return nil, err
		}
		desc := strings.TrimSpace(request.Params.Arguments["description"])

		text := renderAddProvisioner(path, ptype, desc)
		return mcp.NewGetPromptResult("Add a provisioner to a warpgate template", userMessage(text)), nil
	}

	s.AddPrompt(prompt, handler)
}

func renderConvertFromPacker(packerPath, outDir string) string {
	convertArgs := fmt.Sprintf("packer_path=%q", packerPath)
	if outDir != "" {
		convertArgs += fmt.Sprintf(", output=%q", outDir)
	}

	return fmt.Sprintf(`Convert the Packer template at %s into a warpgate template.

Steps:

1. Convert: call the warpgate_convert tool with %s. The CLI will translate
   sources, provisioners, and variables it understands into warpgate.yaml.
2. Review the output: read the generated warpgate.yaml. Flag any of these to
   the user, since warpgate_convert cannot translate them automatically:
   - post-processors with no warpgate equivalent
   - HCL functions inside variable defaults
   - dynamic blocks
   - non-shell/non-ansible provisioners
3. Validate: call warpgate_validate, then warpgate_schema_validate against
   the new template.
4. First build: only after validation passes, call warpgate_build_streaming
   so the user can confirm the converted template builds end-to-end.
5. If the build fails, switch to the debug_failed_build prompt with the new
   template path and the error summary.

Do not delete or modify the original Packer template — leave it in place
until the user has accepted the converted result.`, packerPath, convertArgs)
}

func convertFromPacker(s *server.MCPServer, logger *logging.Logger) {
	prompt := mcp.NewPrompt(
		"convert_from_packer",
		mcp.WithPromptDescription(
			"Convert an existing Packer template into a warpgate template, then validate the result.",
		),
		mcp.WithArgument("packer_path",
			mcp.ArgumentDescription("Path to the existing Packer template (.pkr.hcl or directory)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("output_dir",
			mcp.ArgumentDescription("Output directory for the converted warpgate template"),
		),
	)

	handler := func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		packerPath, err := requireArg(request.Params.Arguments, "packer_path")
		if err != nil {
			logger.Errorf("convert_from_packer: %v", err)
			return nil, err
		}
		outDir := strings.TrimSpace(request.Params.Arguments["output_dir"])

		text := renderConvertFromPacker(packerPath, outDir)
		return mcp.NewGetPromptResult("Convert a Packer template to warpgate", userMessage(text)), nil
	}

	s.AddPrompt(prompt, handler)
}

func renderPublishMultiarchImage(name, registry, tag, archs string) string {
	if tag == "" {
		tag = "latest"
	}
	if archs == "" {
		archs = "amd64,arm64"
	}
	fullRef := fmt.Sprintf("%s/%s:%s", strings.TrimRight(registry, "/"), name, tag)

	return fmt.Sprintf(`Publish %s as a multi-arch image to %s.

Architectures: %s
Final ref: %s

Steps:

1. Pre-flight:
   - Fetch warpgate://config and confirm registry credentials for %s are
     configured. If not, ask the user to populate them before continuing.
   - Confirm with the user that overwriting %s is acceptable.
2. Per-arch build: for each architecture in %s, call warpgate_build_streaming
   with the template name and the arch. Capture the resulting per-arch image
   refs (e.g. %s-amd64, %s-arm64).
3. Manifest create: call warpgate_manifests_create referencing the per-arch
   image refs from step 2 to assemble a multi-arch manifest list under %s.
4. Manifest push: call warpgate_manifests_push to publish the manifest to
   %s.
5. Verify: call warpgate_registry_inspect on %s and confirm the manifest
   list contains an entry per requested architecture.

If any step fails, stop and report which step, the error, and the partial
state on the registry — do not retry destructively.`, name, registry, archs, fullRef,
		registry, fullRef, archs, fullRef, fullRef, fullRef, registry, fullRef)
}

func publishMultiarchImage(s *server.MCPServer, logger *logging.Logger) {
	prompt := mcp.NewPrompt(
		"publish_multiarch_image",
		mcp.WithPromptDescription(
			"Build a warpgate template for multiple architectures, assemble a multi-arch manifest, and push to a registry.",
		),
		mcp.WithArgument("template_name",
			mcp.ArgumentDescription("Name of the template to build and publish"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("registry",
			mcp.ArgumentDescription("Target registry, e.g. ghcr.io/cowdogmoo"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("tag",
			mcp.ArgumentDescription("Image tag (default: latest)"),
		),
		mcp.WithArgument("architectures",
			mcp.ArgumentDescription("Comma-separated arch list (default: amd64,arm64)"),
		),
	)

	handler := func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name, err := requireArg(request.Params.Arguments, "template_name")
		if err != nil {
			logger.Errorf("publish_multiarch_image: %v", err)
			return nil, err
		}
		registry, err := requireArg(request.Params.Arguments, "registry")
		if err != nil {
			logger.Errorf("publish_multiarch_image: %v", err)
			return nil, err
		}
		tag := strings.TrimSpace(request.Params.Arguments["tag"])
		archs := strings.TrimSpace(request.Params.Arguments["architectures"])

		text := renderPublishMultiarchImage(name, registry, tag, archs)
		return mcp.NewGetPromptResult("Publish a multi-arch warpgate image", userMessage(text)), nil
	}

	s.AddPrompt(prompt, handler)
}
