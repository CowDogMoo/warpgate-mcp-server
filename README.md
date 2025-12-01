# Warpgate MCP Server

An MCP (Model Context Protocol) server that provides tools for managing [Warpgate](https://github.com/CowDogMoo/warpgate) templates and workflows. Warpgate is a robust, automatable engine for building security labs, golden images, and multi-architecture containers using modular Packer templates and Taskfile-driven workflows.

## Features

### Template Management
- **List Templates**: View all available Packer templates
- **Get Template Info**: Get detailed information about specific templates
- **Initialize Templates**: Create lockfiles and initialize Packer plugins
- **Validate Templates**: Check templates for syntax and configuration errors
- **Build Templates**: Create Docker images or AWS AMIs from templates

### Workflow Operations
- **List Tasks**: View all available Taskfile tasks
- **Run Tasks**: Execute specific Taskfile tasks with arguments
- **Run Pre-commit**: Execute pre-commit hooks for code quality
- **Run Image Builder**: Simulate GitHub Actions workflows locally with act

### Resources
- Access to Taskfile.yaml configuration
- Template README files
- Repository metadata

## Prerequisites

1. [Go](https://golang.org/dl/) 1.23 or later
2. [Task](https://taskfile.dev/) - Go task runner
3. [Docker](https://www.docker.com/) - For container operations
4. [Packer](https://www.packer.io/) - For template building
5. Warpgate repository cloned locally

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/cowdogmoo/warpgate-mcp-server.git
cd warpgate-mcp-server

# Build and install
make install
```

### Using Docker

```bash
# Build the Docker image
make docker-build

# Or pull from registry (when available)
docker pull ghcr.io/cowdogmoo/warpgate-mcp-server:latest
```

## Usage

### Standalone Binary

```bash
# Run with auto-detected warpgate path
warpgate-mcp-server stdio

# Run with specific warpgate path
warpgate-mcp-server stdio --warpgate-path /path/to/warpgate

# Run with logging
warpgate-mcp-server stdio --log-file /tmp/warpgate-mcp.log
```

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

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

### With Docker

```json
{
  "mcpServers": {
    "warpgate": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "/path/to/warpgate:/warpgate:ro",
        "-v",
        "/var/run/docker.sock:/var/run/docker.sock",
        "warpgate-mcp-server:dev",
        "stdio",
        "--warpgate-path",
        "/warpgate"
      ]
    }
  }
}
```

### With Claude Code

```bash
# Add the server
claude mcp add warpgate -s user -t stdio -- /path/to/warpgate-mcp-server stdio --warpgate-path /path/to/warpgate

# Or with Docker
claude mcp add warpgate -s user -t stdio -- docker run -i --rm -v /path/to/warpgate:/warpgate:ro warpgate-mcp-server:dev stdio --warpgate-path /warpgate
```

## Available Tools

### Template Management

#### `list_templates`
Lists all available Packer templates in the repository.

**Example:**
```json
{
  "templates": ["attack-box", "sliver", "atomic-red-team", "ttpforge", "runzero-explorer"],
  "count": 5,
  "repo_path": "/Users/username/warpgate"
}
```

#### `create_template`
Create a new Packer template with all required files and structure.

**Parameters:**
- `template_name` (string, required): Name of the new template (e.g., 'my-awesome-template')
- `description` (string, required): Brief description of what this template creates
- `base_image` (string, optional): Base Docker image to use (default: 'ubuntu')
- `base_image_version` (string, optional): Version of the base image (default: 'latest')
- `include_ami` (boolean, optional): Include AWS AMI configuration (default: false)

**Example:**
```json
{
  "template_name": "security-toolkit",
  "description": "A comprehensive security testing toolkit with common tools",
  "base_image": "kalilinux/kali-rolling",
  "base_image_version": "latest",
  "include_ami": true
}
```

Creates:
- `plugins.pkr.hcl` - Packer plugin requirements
- `locals.pkr.hcl` - Local variables
- `variables.pkr.hcl` - Template variables
- `docker.pkr.hcl` - Docker build configuration
- `ami.pkr.hcl` - AWS AMI configuration (if include_ami is true)
- `README.md` - Template documentation

#### `get_template_info`
Get detailed information about a specific template.

**Parameters:**
- `template_name` (string, required): Name of the template

**Example:**
```json
{
  "name": "attack-box",
  "path": "/Users/username/warpgate/packer-templates/attack-box",
  "files": ["docker.pkr.hcl", "ami.pkr.hcl", "variables.pkr.hcl", "README.md"],
  "readme": "# Attack Box Template\n..."
}
```

#### `init_template`
Initialize a Packer template (creates lockfiles and initializes plugins).

**Parameters:**
- `template_name` (string, required): Name of the template to initialize

#### `validate_template`
Validate a Packer template for syntax and configuration errors.

**Parameters:**
- `template_name` (string, required): Name of the template to validate

#### `build_template`
Build a Packer template to create Docker images or AWS AMIs.

**Parameters:**
- `template_name` (string, required): Name of the template to build
- `only` (string, optional): Build filter (e.g., 'docker.amd64', 'docker.*', 'amazon-ebs.*')
- `vars` (string, optional): Additional variables in 'key=value key2=value2' format
- `force` (boolean, optional): Force rebuild even if artifacts exist

### Workflow Operations

#### `list_tasks`
List all available Taskfile tasks.

#### `run_task`
Run a specific Taskfile task with optional arguments.

**Parameters:**
- `task_name` (string, required): Name of the task to run
- `args` (object, optional): Arguments to pass as key-value pairs

#### `run_precommit`
Run pre-commit hooks to validate code quality.

#### `run_image_builder`
Run the GitHub Actions image-builder workflow locally using act.

**Parameters:**
- `template` (string, optional): Specific template to build

## Common Workflows

### Create a New Template

```
1. create_template - Create a new template with scaffolding
2. init_template - Initialize Packer plugins
3. validate_template - Check for errors
4. build_template - Build the image
```

### Build an Existing Template

```
1. list_templates - See available templates
2. get_template_info - Learn about a specific template
3. init_template - Initialize the template
4. validate_template - Check for errors
5. build_template - Build the image
```

### Test CI/CD Locally

```
1. run_image_builder - Simulate GitHub Actions locally
```

### Code Quality Check

```
1. run_precommit - Run all pre-commit hooks
```

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Format Code

```bash
make fmt
```

### Clean

```bash
make clean
```

## Environment Variables

- `TASK_X_REMOTE_TASKFILES`: Automatically set to `1` for remote taskfile support
- `PACKER_LOG`: Set to `1` to enable Packer debug logging

## Architecture

```
warpgate-mcp-server/
├── cmd/
│   └── warpgate-mcp-server/    # Main application
│       ├── main.go              # Entry point and server setup
│       └── instructions.md      # MCP server instructions
├── pkg/
│   ├── client/                  # Warpgate client
│   │   └── warpgate.go         # Task execution and repo management
│   ├── logging/                 # Logging utilities
│   │   └── logging.go          # slog-based logger
│   ├── resources/               # MCP resources
│   │   └── resources.go        # Taskfile and template resources
│   └── tools/                   # MCP tools
│       ├── tools.go            # Tool registration
│       ├── list_templates.go   # Template listing
│       ├── template_info.go    # Template information
│       ├── template_operations.go # Init, validate, build
│       └── workflow_operations.go # Task and workflow execution
├── version/
│   └── version.go              # Version information
├── Dockerfile                   # Container image definition
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── README.md                    # This file
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
