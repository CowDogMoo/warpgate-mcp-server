# Warpgate MCP Server

An MCP (Model Context Protocol) server that provides tools for managing
[Warpgate](https://github.com/CowDogMoo/warpgate) templates and workflows.
Warpgate is a pure Go tool for building container images and AWS AMIs,
replacing Packer with a simpler, more integrated workflow.

## Features

### Warpgate CLI Integration

- **Build**: Build container images or AMIs using BuildKit
- **Validate**: Validate warpgate.yaml template configurations
- **Init**: Initialize new templates with scaffolding
- **Convert**: Convert Packer templates to warpgate YAML format

### Template Registry

- **List Templates**: View all templates from local, git, and official sources
- **Template Info**: Get detailed information about templates
- **Add Sources**: Add git repositories or local directories as template sources
- **Remove Sources**: Remove template sources from the registry

### Multi-Architecture Support

- **Create Manifests**: Create multi-arch manifests from architecture-specific builds
- **Push Manifests**: Push manifests to container registries

### Registry Operations

- **List Images**: List available image tags in a container registry
- **Inspect Images**: Inspect container image manifests with architecture details

### Configuration Management

- **Get Config**: Retrieve warpgate configuration values
- **Set Config**: Update warpgate configuration
- **Show Config**: Display current configuration with all settings

### Schema Validation

- **Validate Schema**: Validate warpgate.yaml files against the template schema

### Resources

- Warpgate configuration
- CLI information (version, binary path)
- Template schema, README, and configuration files

## Prerequisites

1. [Go](https://golang.org/dl/) 1.23 or later
2. [Warpgate CLI](https://github.com/CowDogMoo/warpgate) >= 1.0.0
3. [Docker](https://www.docker.com/) - For container operations

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/cowdogmoo/warpgate-mcp-server.git
cd warpgate-mcp-server

# Build
go build -o warpgate-mcp-server ./cmd/warpgate-mcp-server

# Install to Go bin
go install ./cmd/warpgate-mcp-server
```

### Using Docker

```bash
# Build the Docker image
docker build -t warpgate-mcp-server:latest .

# Or pull from registry
docker pull ghcr.io/cowdogmoo/warpgate-mcp-server:latest
```

## Usage

### Standalone Binary

```bash
# Run with auto-detected warpgate path
warpgate-mcp-server stdio

# Run with specific warpgate repository path
warpgate-mcp-server stdio --warpgate-path /path/to/warpgate

# Run with logging
warpgate-mcp-server stdio --log-file /tmp/warpgate-mcp.log
```

### With Claude Desktop

Add to your Claude Desktop configuration
(`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "warpgate": {
      "command": "/path/to/warpgate-mcp-server",
      "args": [
        "stdio",
        "--warpgate-path",
        "/path/to/warpgate"
      ]
    }
  }
}
```

### With Claude Code

```bash
# Add the server
claude mcp add warpgate -s user -t stdio -- /path/to/warpgate-mcp-server stdio --warpgate-path /path/to/warpgate

# Or use the install-mcp task from the warpgate repository
task install-mcp
```

## Available Tools

### Build Tools

#### `warpgate_build`

Build a container image or AMI using the warpgate CLI.

**Parameters:**

- `template` (string, required): Template name, config file path, or warpgate.yaml location
- `target` (string, optional): Build target type ('container' or 'ami')
- `architectures` (array, optional): Target architectures (['amd64', 'arm64'])
- `push` (boolean, optional): Push image to registry after build
- `registry` (string, optional): Container registry to push to
- `vars` (object, optional): Variable overrides as key-value pairs
- `tags` (array, optional): Additional tags to apply
- `no_cache` (boolean, optional): Disable build caching
- `save_digests` (boolean, optional): Save image digests to files

#### `warpgate_validate`

Validate a warpgate template configuration.

**Parameters:**

- `template` (string, required): Path to warpgate.yaml or template directory
- `syntax_only` (boolean, optional): Only validate syntax, skip file existence checks

#### `warpgate_init`

Initialize a new warpgate template with scaffolding.

**Parameters:**

- `name` (string, required): Name of the new template
- `output` (string, optional): Output directory for the template
- `from` (string, optional): Fork from an existing template

### Template Tools

#### `warpgate_templates_list`

List all available templates from configured sources.

**Parameters:**

- `source` (string, optional): Filter by source ('all', 'local', 'git', or specific repo)
- `format` (string, optional): Output format ('table', 'json', 'gha-matrix')

#### `warpgate_templates_info`

Get detailed information about a specific template.

**Parameters:**

- `template` (string, required): Template name

#### `warpgate_templates_add`

Add a template source to the registry.

**Parameters:**

- `source` (string, required): Git URL or local directory path
- `name` (string, optional): Alias for the source

#### `warpgate_templates_remove`

Remove a template source from the registry.

**Parameters:**

- `name` (string, required): Source name or path to remove

#### `create_template`

Create a new warpgate template with warpgate.yaml configuration and scaffolding.

**Parameters:**

- `template_name` (string, required): Name of the new template
- `description` (string, required): Brief description of what this template creates
- `base_image` (string, optional): Base Docker image (default: 'ubuntu')
- `base_image_version` (string, optional): Version of base image (default: '22.04')
- `platforms` (array, optional): Target platforms (default: ['linux/amd64', 'linux/arm64'])
- `include_ami` (boolean, optional): Include AWS AMI target configuration

### Manifest Tools

#### `warpgate_manifests_create`

Create a multi-architecture manifest.

**Parameters:**

- `name` (string, required): Manifest name (e.g., 'registry/image:tag')
- `images` (array, required): List of image references or digest files
- `push` (boolean, optional): Push manifest after creation

#### `warpgate_manifests_push`

Push a manifest to the registry.

**Parameters:**

- `name` (string, required): Manifest name to push
- `purge` (boolean, optional): Purge local manifest after pushing

### Config Tools

#### `warpgate_config_get`

Get warpgate configuration values.

**Parameters:**

- `key` (string, optional): Configuration key (returns all if not specified)

#### `warpgate_config_set`

Set a warpgate configuration value.

**Parameters:**

- `key` (string, required): Configuration key
- `value` (string, required): Value to set

#### `warpgate_config_show`

Show the current warpgate configuration.

### Registry Tools

#### `warpgate_registry_list`

List available image tags in a container registry.

**Parameters:**

- `name` (string, required): Image name (e.g., attack-box, sliver)
- `registry` (string, required): Container registry URL (e.g., ghcr.io/cowdogmoo)
- `namespace` (string, optional): Namespace/organization within the registry
- `auth_file` (string, optional): Path to authentication file

#### `warpgate_registry_inspect`

Inspect a container image manifest from a registry.

**Parameters:**

- `name` (string, required): Image name (e.g., attack-box, sliver)
- `registry` (string, required): Container registry URL (e.g., ghcr.io/cowdogmoo)
- `tags` (array, optional): Image tags to inspect (default: latest)
- `namespace` (string, optional): Namespace/organization within the registry
- `auth_file` (string, optional): Path to authentication file

### Other Tools

#### `warpgate_convert`

Convert a Packer template to warpgate YAML format.

**Parameters:**

- `source` (string, required): Path to Packer template
- `output` (string, optional): Output path for warpgate.yaml

#### `warpgate_schema_validate`

Validate a warpgate.yaml configuration file against the template schema.

**Parameters:**

- `config_path` (string, optional): Path to the warpgate.yaml file to validate
- `template_dir` (string, optional): Path to a template directory containing warpgate.yaml

## Common Workflows

### Build a Container Image

```text
1. warpgate_templates_list - See available templates
2. warpgate_templates_info - Get template details
3. warpgate_validate - Validate the configuration
4. warpgate_build - Build the image
```

### Create a New Template

```text
1. warpgate_init - Create template scaffolding
2. Edit warpgate.yaml to configure provisioning
3. warpgate_validate - Check for errors
4. warpgate_build - Build the image
```

### Multi-Architecture Build

```text
1. warpgate_build with architectures=['amd64'] --save-digests
2. warpgate_build with architectures=['arm64'] --save-digests
3. warpgate_manifests_create with both digests
```

### Migrate from Packer

```text
1. warpgate_convert - Convert Packer template to warpgate.yaml
2. warpgate_validate - Validate the converted template
3. warpgate_build - Build with new format
```

### Inspect Registry Images

```text
1. warpgate_registry_list - List available tags for an image
2. warpgate_registry_inspect - Get manifest details and architectures
```

## MCP Resources

The server provides access to these MCP resources:

| Resource URI | Description |
|--------------|-------------|
| `warpgate://config` | Current warpgate CLI configuration |
| `warpgate://cli-info` | CLI version and binary path information |
| `warpgate://schema/template` | JSON schema for warpgate.yaml validation |
| `warpgate://template/{name}/readme` | Template README documentation |
| `warpgate://template/{name}/config` | Template warpgate.yaml configuration |

## Environment Variables

- `WARPGATE_*` - Warpgate CLI configuration overrides

## Architecture

```text
warpgate-mcp-server/
├── cmd/
│   └── warpgate-mcp-server/       # Main application
│       ├── main.go                 # Entry point and server setup
│       └── instructions.md         # MCP server instructions
├── pkg/
│   ├── client/                     # Warpgate client
│   │   ├── warpgate.go            # CLI detection and execution
│   │   └── mock_client.go         # Mock client for testing
│   ├── logging/                    # Logging utilities
│   │   └── logging.go             # slog-based logger
│   ├── resources/                  # MCP resources
│   │   └── resources.go           # Config, CLI info, schema resources
│   └── tools/                      # MCP tools
│       ├── tools.go               # Tool registration
│       ├── warpgate_build.go      # Build tool
│       ├── warpgate_validate.go   # Validate tool
│       ├── warpgate_init.go       # Init tool
│       ├── warpgate_templates.go  # Template registry tools
│       ├── warpgate_manifests.go  # Manifest tools
│       ├── warpgate_config.go     # Config tools
│       ├── warpgate_convert.go    # Convert tool
│       ├── warpgate_registry.go   # Registry list/inspect tools
│       ├── warpgate_schema.go     # Schema validation tool
│       └── create_template.go     # Template creation tool
├── version/
│   └── version.go                  # Version information
├── Dockerfile                      # Container image definition
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## License

MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

For issues and questions:

- Open an issue on [GitHub](https://github.com/cowdogmoo/warpgate-mcp-server/issues)
- Refer to the [Warpgate documentation](https://github.com/CowDogMoo/warpgate)

## Related Projects

- [Warpgate](https://github.com/CowDogMoo/warpgate) - The main Warpgate project
- [MCP Go SDK](https://github.com/mark3labs/mcp-go) - Go SDK for Model Context Protocol
