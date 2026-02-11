package copamcp

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates and configures the MCP server with all tools
func NewServer(version string) *mcp.Server {
	if version == "" {
		version = "dev"
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "copacetic-mcp",
		Version: version,
	}, nil)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "version",
		Description: "Copacetic automated container patching",
	}, Version)

	// Workflow guidance tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "workflow-guide",
		Description: "Get guidance on which Copacetic tools to use for different container patching scenarios",
	}, WorkflowGuide)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "scan-container",
		Description: "Scan container image for vulnerabilities using Trivy - creates vulnerability reports required for report-based patching",
	}, ScanContainer)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "patch-comprehensive",
		Description: "Comprehensively patch all container image platforms with Copa - patches all available platforms WITHOUT vulnerability scanning. Use ONLY when you want to patch all platforms regardless of vulnerabilities. For vulnerability-based patching, use 'scan-container' + 'patch-vulnerabilities'.",
	}, PatchComprehensive)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "patch-platform-selective",
		Description: "Patch specific container image platforms with Copa - patches only the specified platforms WITHOUT vulnerability scanning. Use ONLY when you want to patch specific platforms regardless of vulnerabilities. For vulnerability-based patching, use 'scan-container' + 'patch-vulnerabilities'.",
	}, PatchPlatformSelective)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "patch-report-based",
		Description: "Patch container image vulnerabilities using a pre-generated vulnerability report from 'scan-container' tool - requires running 'scan-container' first. This is the RECOMMENDED approach for vulnerability-based patching.",
	}, PatchReportBased)

	return server
}

// Run starts the MCP server
func Run(ctx context.Context, version string) error {
	server := NewServer(version)
	return server.Run(ctx, &mcp.StdioTransport{})
}

// getWorkflowGuidance provides guidance on which tool to use for different scenarios
func getWorkflowGuidance() string {
	return `
=== COPACETIC WORKFLOW GUIDANCE ===

Choose the right tool for your use case:

1. VULNERABILITY-BASED PATCHING (Recommended):
   Step 1: scan-container (scan for vulnerabilities)
   Step 2: patch-vulnerabilities (patch only found vulnerabilities)
   
2. PLATFORM-SPECIFIC PATCHING (without vulnerability scanning):
   Use: patch-platforms (specify which platforms to patch)
   
3. COMPREHENSIVE PATCHING (without vulnerability scanning):
   Use: patch-comprehensive (patch all available platforms)

IMPORTANT: Do NOT mix approaches. If you scan first, use patch-vulnerabilities.
If you want platform-specific patching without scanning, use patch-platforms.`
}
