# Contributing to Warpgate MCP Server

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.23 or later
- pre-commit
- Task (go-task)
- Docker (for container builds)
- git

### Getting Started

1. Fork and clone the repository:

```bash
git clone https://github.com/cowdogmoo/warpgate-mcp-server.git
cd warpgate-mcp-server
```

2. Install dependencies:

```bash
go mod download
```

3. Install pre-commit hooks:

```bash
pre-commit install
```

4. Build the project:

```bash
make build
```

5. Run tests:

```bash
make test
```

## Development Workflow

### Code Quality

We use several tools to maintain code quality:

- **pre-commit hooks**: Automatically run before each commit
- **golangci-lint**: Go linter aggregator
- **gofmt/goimports**: Code formatting
- **go vet**: Static analysis
- **govulncheck**: Vulnerability scanning
- **semgrep**: Security analysis

### Running Pre-commit Hooks

Pre-commit hooks run automatically on `git commit`, but you can also run them manually:

```bash
# Run on all files
pre-commit run --all-files

# Run on specific files
pre-commit run --files <file1> <file2>
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
bash .hooks/run-go-tests.sh coverage

# Run tests for modified files
bash .hooks/run-go-tests.sh modified
```

### Code Formatting

```bash
# Format code
make fmt

# Check imports
go run golang.org/x/tools/cmd/goimports -w .
```

### Linting

```bash
# Run golangci-lint
make lint

# Or directly
golangci-lint run ./...
```

## Pull Request Process

1. **Create a feature branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clear, concise commit messages
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests and checks**:

   ```bash
   make test
   pre-commit run --all-files
   ```

4. **Commit your changes**:

   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

5. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**:
   - Provide a clear description of the changes
   - Reference any related issues
   - Ensure all CI checks pass

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Examples:

```
feat: add support for multi-region builds
fix: correct template path resolution
docs: update installation instructions
```

## Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise
- Write tests for new functionality

## Testing

- Write unit tests for new code
- Maintain or improve code coverage
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example test structure:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported types and functions
- Update API documentation if applicable
- Include examples in documentation

## Release Process

Releases are automated using GoReleaser:

1. Update version in `version/version.go`
2. Create and push a tag:

   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

3. GitHub Actions will automatically build and publish the release

## Getting Help

- Open an issue for bugs or feature requests
- Join discussions in GitHub Discussions
- Check existing issues and PRs before creating new ones

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help others learn and grow
- Follow the project's code of conduct

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
