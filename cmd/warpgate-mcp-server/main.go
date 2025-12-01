// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	_ "embed"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/resources"
	"github.com/cowdogmoo/warpgate-mcp-server/pkg/tools"
	"github.com/cowdogmoo/warpgate-mcp-server/version"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

//go:embed instructions.md
var instructions string

func runStdioServer(logger *logging.Logger, warpgatePath string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mcpServer := NewServer(version.Version, logger, warpgatePath)
	return serverInit(ctx, mcpServer, logger)
}

func NewServer(ver string, logger *logging.Logger, warpgatePath string) *server.MCPServer {
	// Create a new MCP server with capabilities
	s := server.NewMCPServer(
		"warpgate-mcp-server",
		ver,
		server.WithResourceCapabilities(true, false),
	)

	// Register tools and resources
	tools.RegisterTools(s, logger, warpgatePath)
	resources.RegisterResources(s, logger, warpgatePath)

	return s
}

// runDefaultCommand handles the default behavior when no subcommand is provided
func runDefaultCommand(cmd *cobra.Command, _ []string) {
	logFile, err := cmd.PersistentFlags().GetString("log-file")
	if err != nil {
		stdlog.Fatal("Failed to get log file:", err)
	}

	warpgatePath, err := cmd.PersistentFlags().GetString("warpgate-path")
	if err != nil {
		stdlog.Fatal("Failed to get warpgate path:", err)
	}

	logger, err := logging.NewLogger(logFile)
	if err != nil {
		stdlog.Fatal("Failed to initialize logger:", err)
	}

	if err := runStdioServer(logger, warpgatePath); err != nil {
		stdlog.Fatal("failed to run stdio server:", err)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:     "warpgate-mcp-server",
		Short:   "Warpgate MCP Server",
		Long:    `An MCP server that provides tools for managing Warpgate templates and workflows.`,
		Version: version.GetFullVersion(),
		Run:     runDefaultCommand,
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		Run: func(cmd *cobra.Command, _ []string) {
			logFile, err := rootCmd.PersistentFlags().GetString("log-file")
			if err != nil {
				stdlog.Fatal("Failed to get log file:", err)
			}

			warpgatePath, err := rootCmd.PersistentFlags().GetString("warpgate-path")
			if err != nil {
				stdlog.Fatal("Failed to get warpgate path:", err)
			}

			logger, err := logging.NewLogger(logFile)
			if err != nil {
				stdlog.Fatal("Failed to initialize logger:", err)
			}

			if err := runStdioServer(logger, warpgatePath); err != nil {
				stdlog.Fatal("failed to run stdio server:", err)
			}
		},
	}
)

func init() {
	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
	rootCmd.PersistentFlags().String("warpgate-path", "", "Path to warpgate repository (default: auto-detect)")

	rootCmd.AddCommand(stdioCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func serverInit(ctx context.Context, mcpServer *server.MCPServer, logger *logging.Logger) error {
	stdioServer := server.NewStdioServer(mcpServer)
	stdLogger := stdlog.New(logger.Writer(), "stdioserver", 0)
	stdioServer.SetErrorLogger(stdLogger)

	// Start listening for messages
	errC := make(chan error, 1)
	go func() {
		errC <- stdioServer.Listen(ctx, os.Stdin, os.Stdout)
	}()

	_, _ = fmt.Fprintf(os.Stderr, "Warpgate MCP Server running on stdio\n")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		logger.Infof("shutting down server...")
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("error running server: %w", err)
		}
	}

	return nil
}
