package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	// "path"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Global variables for MCP client session
var (
	client  *mcp.Client
	session *mcp.ClientSession
	ctx     context.Context
)

func executeMCPTool(toolName string, args map[string]any) error {
	fmt.Printf("\n=== Executing %s tool ===\n", toolName)

	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	}

	res, err := session.CallTool(ctx, params)
	if err != nil {
		return fmt.Errorf("CallTool failed for %s: %v", toolName, err)
	}

	if res.IsError {
		for _, c := range res.Content {
			if text, ok := c.(*mcp.TextContent); ok {
				return fmt.Errorf("%s tool failed: %s", toolName, text.Text)
			}
		}
		return fmt.Errorf("%s tool failed with unknown error", toolName)
	}

	for _, c := range res.Content {
		fmt.Printf("Result: %s\n", c.(*mcp.TextContent).Text)
	}
	return nil
}

func initMCPClient() error {
	ctx = context.Background()

	client = mcp.NewClient(
		&mcp.Implementation{Name: "copamcp-cli", Version: "v1.0.0"},
		&mcp.ClientOptions{
			LoggingMessageHandler: func(ctx context.Context, req *mcp.LoggingMessageRequest) {
				fmt.Printf("[server log][%s] %v\n", req.Params.Level, req.Params.Data)
			},
		},
	)

	// Try to find the server binary in common locations
	serverPaths := []string{
		"./bin/copacetic-mcp-server",
		"../bin/copacetic-mcp-server",
		"./copacetic-mcp-server",
		"copacetic-mcp-server",
	}

	var serverPath string
	for _, path := range serverPaths {
		if _, err := os.Stat(path); err == nil {
			serverPath = path
			break
		}
	}

	if serverPath == "" {
		return fmt.Errorf("could not find copacetic-mcp-server binary in any of: %v", serverPaths)
	}

	cmd := exec.Command(serverPath, "stdio")
	// Capture server's stderr for logging
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}
	// Start goroutine to read and log stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			log.Printf("[server stderr] %s", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading server stderr: %v", err)
		}
	}()

	transport := &mcp.CommandTransport{Command: cmd}
	session, err = client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %v", err)
	}

	// Enable receiving log messages from the server
	err = session.SetLoggingLevel(ctx, &mcp.SetLoggingLevelParams{Level: "debug"})
	if err != nil {
		log.Printf("Warning: failed to set logging level: %v", err)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "copamcp",
		Short: "Copacetic MCP CLI - A command-line interface for container patching",
		Long: `Copacetic MCP CLI provides a command-line interface for the Copacetic MCP server,
allowing you to patch container images, scan for vulnerabilities, and manage container security.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := initMCPClient(); err != nil {
				log.Fatalf("Failed to initialize MCP client: %v", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if session != nil {
				session.Close()
			}
		},
	}

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Get Copacetic version information",
		Long:  "Retrieve version information from the Copacetic tool",
		Run: func(cmd *cobra.Command, args []string) {
			if err := executeMCPTool("version", map[string]any{}); err != nil {
				log.Fatalf("Error executing version command: %v", err)
			}
		},
	}

	// Scan command
	var (
		scanImage     string
		scanPlatforms []string
	)
	var scanCmd = &cobra.Command{
		Use:   "scan-container",
		Short: "Scan container image for vulnerabilities",
		Long:  "Scan a container image for vulnerabilities using Trivy",
		Run: func(cmd *cobra.Command, args []string) {
			mcpArgs := map[string]any{
				"image": scanImage,
			}
			if len(scanPlatforms) > 0 {
				mcpArgs["platform"] = scanPlatforms
			}
			if err := executeMCPTool("scan-container", mcpArgs); err != nil {
				log.Fatalf("Error executing scan-container command: %v", err)
			}
		},
	}
	scanCmd.Flags().StringVarP(&scanImage, "image", "i", "", "Container image to scan (required)")
	scanCmd.Flags().StringSliceVarP(&scanPlatforms, "platform", "p", []string{}, "Target platform(s) for scanning (e.g., linux/amd64,linux/arm64)")
	scanCmd.MarkFlagRequired("image")

	// Patch Comprehensive command
	var (
		comprehensiveImage    string
		comprehensivePatchTag string
		comprehensivePush     bool
	)
	var patchComprehensiveCmd = &cobra.Command{
		Use:   "patch-comprehensive",
		Short: "Patch all container image platforms",
		Long:  "Patch all available platforms in a container image without vulnerability scanning",
		Run: func(cmd *cobra.Command, args []string) {
			mcpArgs := map[string]any{
				"image":    comprehensiveImage,
				"patchtag": comprehensivePatchTag,
				"push":     comprehensivePush,
			}
			if err := executeMCPTool("patch-comprehensive", mcpArgs); err != nil {
				log.Fatalf("Error executing patch-comprehensive command: %v", err)
			}
		},
	}
	patchComprehensiveCmd.Flags().StringVarP(&comprehensiveImage, "image", "i", "", "Container image to patch (required)")
	patchComprehensiveCmd.Flags().StringVarP(&comprehensivePatchTag, "patchtag", "t", "", "Tag for the patched image")
	patchComprehensiveCmd.Flags().BoolVarP(&comprehensivePush, "push", "", false, "Push patched image to registry")
	patchComprehensiveCmd.MarkFlagRequired("image")
	// patchComprehensiveCmd.MarkFlagRequired("patchtag")

	// Patch Platforms command
	var (
		platformsImage    string
		platformsPatchTag string
		platformsPush     bool
		targetPlatforms   []string
	)
	var patchPlatformsCmd = &cobra.Command{
		Use:   "patch-platforms",
		Short: "Patch specific container image platforms",
		Long:  "Patch only specified platforms in a container image without vulnerability scanning",
		Run: func(cmd *cobra.Command, args []string) {
			mcpArgs := map[string]any{
				"image":    platformsImage,
				"patchtag": platformsPatchTag,
				"push":     platformsPush,
				"platform": targetPlatforms,
			}
			if err := executeMCPTool("patch-platform-selective", mcpArgs); err != nil {
				log.Fatalf("Error executing patch-platforms command: %v", err)
			}
		},
	}
	patchPlatformsCmd.Flags().StringVarP(&platformsImage, "image", "i", "", "Container image to patch (required)")
	patchPlatformsCmd.Flags().StringVarP(&platformsPatchTag, "patchtag", "t", "", "Tag for the patched image (required)")
	patchPlatformsCmd.Flags().BoolVarP(&platformsPush, "push", "", false, "Push patched image to registry")
	patchPlatformsCmd.Flags().StringSliceVarP(&targetPlatforms, "platform", "p", []string{}, "Target platform(s) for patching (required)")
	patchPlatformsCmd.MarkFlagRequired("image")
	patchPlatformsCmd.MarkFlagRequired("patchtag")
	patchPlatformsCmd.MarkFlagRequired("platform")

	// Patch Vulnerabilities command
	var (
		vulnImage      string
		vulnPatchTag   string
		vulnPush       bool
		vulnReportPath string
	)
	var patchVulnerabilitiesCmd = &cobra.Command{
		Use:   "patch-vulnerabilities",
		Short: "Patch container image vulnerabilities",
		Long:  "Patch container image vulnerabilities using a pre-generated vulnerability report from scan command",
		Run: func(cmd *cobra.Command, args []string) {
			mcpArgs := map[string]any{
				"image":      vulnImage,
				"patchtag":   vulnPatchTag,
				"push":       vulnPush,
				"reportPath": vulnReportPath,
			}
			if err := executeMCPTool("patch-report-based", mcpArgs); err != nil {
				log.Fatalf("Error executing patch-vulnerabilities command: %v", err)
			}
		},
	}
	patchVulnerabilitiesCmd.Flags().StringVarP(&vulnImage, "image", "i", "", "Container image to patch (required)")
	patchVulnerabilitiesCmd.Flags().StringVarP(&vulnPatchTag, "patchtag", "t", "", "Tag for the patched image (required)")
	patchVulnerabilitiesCmd.Flags().BoolVarP(&vulnPush, "push", "", false, "Push patched image to registry")
	patchVulnerabilitiesCmd.Flags().StringVarP(&vulnReportPath, "report-path", "r", "", "Path to vulnerability report directory (required)")
	patchVulnerabilitiesCmd.MarkFlagRequired("image")
	patchVulnerabilitiesCmd.MarkFlagRequired("patchtag")
	patchVulnerabilitiesCmd.MarkFlagRequired("report-path")

	// List tools command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available MCP tools",
		Long:  "List all available MCP tools from the server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("\n=== Available MCP Tools ===")
			listRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
			if err != nil {
				log.Fatalf("Failed to list tools: %v", err)
			}
			for _, tool := range listRes.Tools {
				fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
			}
		},
	}

	// Add all commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(patchComprehensiveCmd)
	rootCmd.AddCommand(patchPlatformsCmd)
	rootCmd.AddCommand(patchVulnerabilitiesCmd)
	rootCmd.AddCommand(listCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
