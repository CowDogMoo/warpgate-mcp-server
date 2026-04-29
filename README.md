# Warpgate MCP Server

MCP server for [warpgate](https://github.com/CowDogMoo/warpgate) - a Go-native tool for building multi-arch containers and AMIs.

## Prerequisites

- [warpgate](https://github.com/cowdogmoo/warpgate) CLI installed and in PATH
- Go 1.23+ (for building from source)

## Installation

```bash
git clone https://github.com/cowdogmoo/warpgate-mcp-server.git
cd warpgate-mcp-server
go build -o warpgate-mcp-server ./cmd/warpgate-mcp-server
```

## Usage

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "warpgate": {
      "command": "/path/to/warpgate-mcp-server",
      "args": ["stdio"]
    }
  }
}
```

### Claude Code

```bash
claude mcp add warpgate -s user -t stdio -- /path/to/warpgate-mcp-server stdio
```

## Tools

### Templates
- `list_templates` — list templates from configured sources
- `search_templates` — substring search by name/description/author/tags
- `get_template_info` — show details for a template
- `init_template` — scaffold a new template
- `validate_template` — validate a warpgate.yaml
- `add_template_source` — register a git repo or local path
- `remove_template_source` — unregister a source
- `update_template_cache` — pull latest from configured git sources

### Build
- `build_template` — build container or AMI (full flag surface: archs, push, push-digest, regions, copy-to-regions, build-args, labels, cache-from/to, output-manifest, dry-run, etc.)
- `convert_packer_template` — migrate a Packer HCL template to warpgate.yaml

### Manifests (multi-arch)
- `create_manifest` — assemble and push a multi-arch manifest from per-arch digest files
- `inspect_manifest` — inspect a multi-arch manifest in a registry
- `list_manifests` — list manifest tags for an image

### AWS
- `cleanup_aws_resources` — clean up AWS Image Builder resources (use `dry_run` first)

### Configuration
- `show_config` — show resolved warpgate configuration
- `get_config` — read a config value by dotted key
- `set_config` — write a config value to `~/.config/warpgate/config.yaml`

## Resources

- `warpgate://config` - Your warpgate configuration
- `warpgate://schema/template` - Template JSON schema
- `warpgate://examples/template` - Example template

## Template Format

Templates use `warpgate.yaml` instead of Packer HCL:

```yaml
metadata:
  name: my-template
  version: 1.0.0

name: my-app
version: latest

base:
  image: ubuntu:22.04

provisioners:
  - type: shell
    inline:
      - apt-get update
      - apt-get install -y curl

targets:
  - type: container
    registry: ghcr.io/myorg
    tags: [latest]
    platforms: [linux/amd64, linux/arm64]
    push: false
```

See `examples/warpgate.yaml` for a complete example.

## License

MIT
