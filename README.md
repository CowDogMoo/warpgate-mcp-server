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

### Template Management
- `list_templates` - List templates from all sources
- `get_template_info` - Get template details
- `init_template` - Create new template
- `validate_template` - Validate warpgate.yaml
- `build_template` - Build container/AMI

### Discovery & Sources
- `search_templates` - Fuzzy search templates
- `add_template_source` - Add git repo or local path
- `remove_template_source` - Remove source
- `update_template_cache` - Refresh cache

### Advanced
- `convert_packer_template` - Migrate from Packer
- `create_manifest` - Multi-arch manifests

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
