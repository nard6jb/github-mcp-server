// Package main is the entry point for the GitHub MCP Server.
// It initializes and starts the Model Context Protocol server that
// provides tools for interacting with the GitHub API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		transport string
		port      int
		logLevel  string
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server — Model Context Protocol server for GitHub",
		Long: `A Model Context Protocol (MCP) server that exposes GitHub API
functionality as tools, resources, and prompts for use with AI assistants.`,
		Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), transport, port, logLevel)
		},
	}

	cmd.Flags().StringVarP(&transport, "transport", "t", "stdio",
		"Transport to use: stdio or sse")
	cmd.Flags().IntVarP(&port, "port", "p", 8080,
		"Port to listen on (only used with sse transport)")
	// Default to debug level for easier local development and troubleshooting.
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", "debug",
		"Log level: debug, info, warn, error")

	return cmd
}

func runServer(ctx context.Context, transport string, port int, logLevel string) error {
	// Retrieve the GitHub token from the environment.
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token == "" {
		fmt.Fprintln(os.Stderr, "Warning: no GitHub token found; set GITHUB_PERSONAL_ACCESS_TOKEN or GITHUB_TOKEN")
	}

	// Set up context that cancels on OS signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := server.Config{
		Transport: transport,
		Port:      port,
		LogLevel:  logLevel,
		Token:     token,
		Version:   Version,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Starting GitHub MCP Server %s (transport: %s)\n", Version, transport)

	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
