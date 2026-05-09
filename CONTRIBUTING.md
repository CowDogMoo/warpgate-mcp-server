# Contributing

Thanks for your interest in contributing to the Warpgate MCP Server.

## Development Setup

### Prerequisites

| Requirement      | Version | Notes                                  |
| ---------------- | ------- | -------------------------------------- |
| **Go**           | 1.24+   | Toolchain for building and testing     |
| **Task**         | 3.x     | [go-task](https://taskfile.dev/) runner |
| **pre-commit**   | 3.x+    | Drives format/lint/security hooks      |
| **Docker**       | 20.10+  | Required for container builds and the Docker target |
| **Warpgate CLI** | 1.0.0+  | The MCP server shells out to `warpgate` |

### Getting Started

```bash
# 1. Fork and clone
git clone https://github.com/<your-fork>/warpgate-mcp-server.git
cd warpgate-mcp-server

# 2. Install dependencies
go mod download

# 3. Install pre-commit hooks
pre-commit install

# 4. Build and run tests
task build
task test
```

## Common Tasks

```bash
task                 # default: install go deps + run pre-commit hooks
task build           # build the binary with version ldflags
task install         # install to $GOPATH/bin
task test            # go test -race ./...
task test-coverage   # coverage report (HTML)
task lint            # golangci-lint run ./...
task fmt             # go fmt ./...
task tidy            # go mod tidy
task docker-build    # build the local Docker image
task install-mcp     # configure Claude Desktop to use this server
```

## Pull Request Workflow

1. Branch from `main`:

   ```bash
   git checkout -b feat/my-feature
   ```

2. Make focused changes. Add or update tests for behavior changes.
3. Run the full local check before pushing:

   ```bash
   task test
   pre-commit run --all-files
   ```

4. Use [Conventional Commits](https://www.conventionalcommits.org/) for
   commit messages (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
5. Open the PR and link any related issues. CI must be green before merge.

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`, `golangci-lint`).
- Add doc comments on exported types and functions.
- Prefer table-driven tests for handlers and pure helpers.
- Keep tools and prompts focused — one file per tool, one render function
  per prompt so behavior can be unit-tested without the MCP server.

## Adding a New Tool

1. Create `pkg/tools/<name>.go` and define a registration function
   (`func myTool(s *server.MCPServer, logger *logging.Logger, warpgatePath string)`).
2. Register the tool from `pkg/tools/tools.go`.
3. If the tool calls the Warpgate CLI, extend the client in
   `pkg/client/warpgate.go` with a strongly-typed wrapper rather than
   shelling out from the handler directly.
4. Add table-driven tests under `pkg/tools/<name>_test.go`.
5. Update the tools table in `README.md`.

## Adding a New Prompt

1. Add a `render<Name>` pure function and a registration function in
   `pkg/prompts/prompts.go`.
2. Register from `RegisterPrompts`.
3. Add a unit test that asserts the rendered text contains the key tool
   names and parameter substitutions.
4. Update the prompts table in `README.md`.

## Releases

Releases are automated by [GoReleaser](https://goreleaser.com/) on tag push.

```bash
git tag -a v0.3.0 -m "Release v0.3.0"
git push origin v0.3.0
```

GitHub Actions builds binaries, container images, and the GitHub Release.

## Reporting Issues

Open an issue on
[GitHub](https://github.com/CowDogMoo/warpgate-mcp-server/issues) with:

- A short description of the problem
- Steps to reproduce (template name, tool call, MCP client)
- Output of `warpgate-mcp-server --version` and `warpgate version`
- Relevant lines from the `--log-file` output (with secrets redacted)

## License

By contributing, you agree your contributions are licensed under the MIT
License.
