package main

import (
	"context"
	"fmt"
	"os"

	"github.com/project-copacetic/mcp-server/internal/copamcp"
	"github.com/spf13/cobra"
)

// These variables are set by the build process using ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "copa-mcp-server",
	Short: "Copacetic MCP Server",
	Long: `A Model Context Protocol (MCP) server for automated container image patching using Copacetic and Trivy.
This server exposes container patching capabilities through the MCP protocol, allowing AI agents and tools to patch container image vulnerabilities programmatically.`,
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date),
}

var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Start stdio server",
	Long:  `Start a server that communicates via standard input/output streams using the Model Context Protocol (MCP).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copamcp.Run(context.Background(), version)
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(stdioCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
