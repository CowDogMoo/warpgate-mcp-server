# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'github.com/cowdogmoo/warpgate-mcp-server/version.GitCommit=$(git rev-parse HEAD)' \
    -X 'github.com/cowdogmoo/warpgate-mcp-server/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o warpgate-mcp-server \
    ./cmd/warpgate-mcp-server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
# task: Go task runner for Taskfile execution
# docker-cli: For docker operations
# git: For git operations
# bash: For shell scripts
RUN apk add --no-cache ca-certificates tzdata task docker-cli git bash

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/warpgate-mcp-server /app/warpgate-mcp-server

# Set the entrypoint
ENTRYPOINT ["/app/warpgate-mcp-server"]
CMD ["stdio"]
