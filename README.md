# Warpgate MCP Server

**Drive [Warpgate](https://github.com/CowDogMoo/warpgate) from any MCP-aware
agent.**

[![License](https://img.shields.io/github/license/CowDogMoo/warpgate-mcp-server?label=License&style=flat&color=blue&logo=github)](https://github.com/CowDogMoo/warpgate-mcp-server/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/CowDogMoo/warpgate-mcp-server?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/CowDogMoo/warpgate-mcp-server?label=Release&logo=github)](https://github.com/CowDogMoo/warpgate-mcp-server/releases)

[![Tests](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/tests.yaml/badge.svg)](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/tests.yaml)
[![Pre-Commit](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/pre-commit.yaml/badge.svg)](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/pre-commit.yaml)
[![Semgrep](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/semgrep.yaml)
[![Renovate](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/renovate.yaml/badge.svg)](https://github.com/CowDogMoo/warpgate-mcp-server/actions/workflows/renovate.yaml)

---

## Overview

A [Model Context Protocol](https://modelcontextprotocol.io) server that wraps
the [Warpgate](https://github.com/CowDogMoo/warpgate) CLI so Claude Code,
Claude Desktop, Cursor, Continue, and other MCP clients can build container
images and AWS AMIs, manage template repositories, and operate on container
registries — without operators leaving the chat.

- Build OCI images and AMIs from YAML templates with BuildKit
- Stream live build progress through MCP logging notifications
- Manage template registries (local, git, official) end-to-end
- Create, push, copy, and delete container images and multi-arch manifests
- Validate `warpgate.yaml` against the embedded JSON schema
- Convert legacy Packer templates to Warpgate format
- Ships with reusable prompt workflows for common operator recipes

**Built for:**

- Security teams building attack/defense imagery
- Platform engineers standardizing base images across teams
- Operators who'd rather describe a pipeline than click through one

## Prerequisites

| Requirement       | Version | Notes                                              |
| ----------------- | ------- | -------------------------------------------------- |
| **Go**            | 1.24+   | Required for `go install`                          |
| **Warpgate CLI**  | 1.0.0+  | The MCP server shells out to `warpgate`            |
| **Docker**        | 20.10+  | Required for container builds                      |
| **Docker Buildx** | 0.8+    | Required for multi-arch builds                     |
| **AWS CLI** (opt) | 2.x     | Required for AMI builds                            |
| **skopeo/crane**  | -       | Required for `warpgate_registry_{copy,delete}`     |

Install Warpgate first:

```bash
go install github.com/cowdogmoo/warpgate/v3/cmd/warpgate@latest
warpgate version
```

## Quick Start

```bash
# Install the MCP server
go install github.com/cowdogmoo/warpgate-mcp-server/cmd/warpgate-mcp-server@latest

# Run over stdio (auto-detects warpgate on PATH)
warpgate-mcp-server stdio

# Run with an explicit warpgate repo + log file
warpgate-mcp-server stdio \
  --warpgate-path "$HOME/cowdogmoo/warpgate" \
  --log-file /tmp/warpgate-mcp.log
```

## Client Setup

### Claude Code

```bash
claude mcp add warpgate -s user -t stdio -- \
  warpgate-mcp-server stdio --warpgate-path "$HOME/cowdogmoo/warpgate"
```

Or, from a checked-out warpgate repository:

```bash
task install-mcp
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)
or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "warpgate": {
      "command": "/path/to/warpgate-mcp-server",
      "args": ["stdio", "--warpgate-path", "/path/to/warpgate"]
    }
  }
}
```

### Docker

```bash
docker run -i --rm \
  -v "$HOME/cowdogmoo/warpgate:/warpgate:ro" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/cowdogmoo/warpgate-mcp-server:latest \
  stdio --warpgate-path /warpgate
```

A canonical [`server.json`](server.json) descriptor is included for
registry-based discovery.

## Tools

### Build & Validate

| Tool                       | Description                                                              |
| -------------------------- | ------------------------------------------------------------------------ |
| `warpgate_build`           | Build a container image or AMI; returns full output when complete        |
| `warpgate_build_streaming` | Build with line-by-line progress via MCP logging notifications           |
| `warpgate_validate`        | Validate a template directory or `warpgate.yaml` via the CLI             |
| `warpgate_schema_validate` | Validate `warpgate.yaml` against the embedded JSON schema (no CLI call)  |
| `warpgate_init`            | Scaffold a new template using the `warpgate init` CLI                    |
| `create_template`          | Generate a new template directory with `warpgate.yaml`, README, scripts  |
| `warpgate_convert`         | Convert a Packer template to `warpgate.yaml`                             |

### Template Registry

| Tool                        | Description                                              |
| --------------------------- | -------------------------------------------------------- |
| `warpgate_templates_list`   | List templates from local, git, and official sources     |
| `warpgate_templates_info`   | Show metadata, provisioners, and computed values         |
| `warpgate_templates_add`    | Register a git URL or local directory as a source        |
| `warpgate_templates_remove` | Remove a registered source                               |

### Manifests & Registries

| Tool                       | Description                                                       |
| -------------------------- | ----------------------------------------------------------------- |
| `warpgate_manifests_create`| Assemble a multi-arch manifest list from per-arch refs/digests    |
| `warpgate_manifests_push`  | Push a manifest list to its registry                              |
| `warpgate_registry_list`   | List image tags in a registry                                     |
| `warpgate_registry_inspect`| Inspect manifests, including per-architecture entries             |
| `warpgate_registry_copy`   | Copy images between registries (skopeo/crane)                     |
| `warpgate_registry_delete` | Delete tags from a registry, with `dry_run` support               |

### Config

| Tool                  | Description                                |
| --------------------- | ------------------------------------------ |
| `warpgate_config_get` | Read a config key (or all keys)            |
| `warpgate_config_set` | Set a config key                           |
| `warpgate_config_show`| Print the merged effective configuration   |

## Prompts

Prompts ship with the server as parameterized workflow recipes. MCP clients
list them automatically; pick one and the server returns a step-by-step
script that chains the tools above.

| Prompt                    | Purpose                                                              |
| ------------------------- | -------------------------------------------------------------------- |
| `bootstrap_new_template`  | Scaffold, configure, validate, and run the first build of a template |
| `debug_failed_build`      | Walk a failing build through validation, schema, config, and logs    |
| `add_provisioner`         | Add a shell, ansible, or file provisioner to an existing template    |
| `convert_from_packer`     | Convert a Packer template and verify it builds end-to-end            |
| `publish_multiarch_image` | Build per-arch, assemble a manifest, push, and verify on the registry |

## Resources

| URI                                  | Description                                              |
| ------------------------------------ | -------------------------------------------------------- |
| `warpgate://config`                  | Effective `warpgate config show` output                  |
| `warpgate://cli-info`                | Detected binary path, version, and minimum required version |
| `warpgate://schema/template`         | JSON Schema for `warpgate.yaml` validation               |
| `warpgate://template/{name}/readme`  | README for a specific template                           |
| `warpgate://template/{name}/config`  | `warpgate templates info` output for a template          |

## Configuration

The MCP server itself takes only two flags; everything else is forwarded to
the Warpgate CLI, which honors its own config file and `WARPGATE_*` env vars.

### Server Flags

| Flag               | Description                                                          |
| ------------------ | -------------------------------------------------------------------- |
| `--warpgate-path`  | Path to a checked-out warpgate repo; used as CWD for CLI invocations |
| `--log-file`       | Path to write MCP server logs (stderr is reserved for the transport) |

### Forwarded Environment

These pass straight through to the Warpgate CLI; refer to the
[Warpgate CLI configuration guide](https://github.com/CowDogMoo/warpgate/blob/main/docs/cli-configuration.md)
for the full list.

| Variable                      | Description                       | Default   |
| ----------------------------- | --------------------------------- | --------- |
| `WARPGATE_LOG_LEVEL`          | Log verbosity (debug/info/...)    | `info`    |
| `WARPGATE_LOG_FORMAT`         | Log format (text, json, color)    | `color`   |
| `WARPGATE_REGISTRY_DEFAULT`   | Default container registry        | `ghcr.io` |
| `WARPGATE_BUILD_DEFAULT_ARCH` | Default build architectures       | `amd64`   |
| `AWS_REGION`                  | AWS region for AMI builds         | -         |
| `AWS_PROFILE`                 | AWS credentials profile           | -         |

## Development

### Build & Test

```bash
# Build with version metadata baked in
task build

# Install to GOPATH/bin
task install

# Run the full test suite with race detector
task test

# Coverage report (HTML)
task test-coverage

# Lint and format
task lint
task fmt

# Run pre-commit hooks (matches CI)
task run-pre-commit

# Build a local Docker image
task docker-build
```

### Project Layout

```text
warpgate-mcp-server/
├── cmd/warpgate-mcp-server/   Entry point + embedded MCP instructions
├── pkg/
│   ├── client/                Warpgate CLI detection and execution
│   ├── logging/               slog-based logger that respects MCP stdio
│   ├── prompts/               Reusable prompt workflows
│   ├── resources/             MCP resource handlers (config, schema, ...)
│   └── tools/                 MCP tool handlers (build, registry, ...)
├── version/                   Version metadata (set via ldflags)
├── server.json                MCP registry descriptor
├── Dockerfile                 Multi-stage container build
└── Taskfile.yaml              Task runner for build/test/release
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, commit
conventions, and the pull request workflow.

## Built With

- [mcp-go](https://github.com/mark3labs/mcp-go) — Go SDK for the Model Context Protocol
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Warpgate](https://github.com/CowDogMoo/warpgate) — The CLI this server fronts

## License

[MIT](LICENSE) — see the LICENSE file for details.
