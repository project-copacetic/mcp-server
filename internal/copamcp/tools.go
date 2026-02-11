package copamcp

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/project-copacetic/mcp-server/internal/copa"
	"github.com/project-copacetic/mcp-server/internal/trivy"
	"github.com/project-copacetic/mcp-server/internal/types"
)

const (
	dryRun = false
)

// PatchComprehensive performs comprehensive patching of all available platforms
// NOTE: This tool patches ALL available platforms WITHOUT vulnerability scanning
// If you want to patch based on vulnerability scan results, use 'scan-container' followed by 'patch-vulnerabilities' instead
func PatchComprehensive(ctx context.Context, req *mcp.CallToolRequest, params types.ComprehensivePatchParams) (*mcp.CallToolResult, any, error) {
	copa := copa.New(params, dryRun)
	_, err := copa.
		Build().
		Run(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("patching failed: %w", err)
	}

	successMsg := fmt.Sprintf("successful patched: %s", params.Image)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: successMsg}},
	}, nil, nil
}

// PatchPlatforms performs platform-selective patching
// NOTE: This tool should only be used when NO vulnerability scanning is desired and specific platforms need patching
// If you want to patch based on vulnerability scan results, use 'patch-vulnerabilities' instead
func PatchPlatformSelective(ctx context.Context, req *mcp.CallToolRequest, params types.PlatformSelectivePatchParams) (*mcp.CallToolResult, any, error) {

	copa := copa.New(params, dryRun)
	_, err := copa.
		BuildWithPlatforms().
		Run(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("platform patch failed: %w", err)
	}

	successMsg := fmt.Sprintf("successful patched: %s", params.Image)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: successMsg}},
	}, nil, nil
}

// PatchVulnerabilities performs report-based patching using an existing vulnerability report
// NOTE: This tool requires that 'scan-container' has been run first to generate the vulnerability report
func PatchReportBased(ctx context.Context, req *mcp.CallToolRequest, params types.ReportBasedPatchParams) (*mcp.CallToolResult, any, error) {
	copa := copa.New(params, dryRun)
	result, err := copa.
		BuildWithReport().
		Run(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("patching failed: %w", err)
	}

	successMsg := fmt.Sprintf("successful patched: %s\n vulnerabilities fixed: %d packages updated: %d", params.Image, result.FixedVulnerabilityCount, result.UpdatedPackageCount)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: successMsg}},
	}, nil, nil
}

// ScanContainer performs vulnerability scanning on a container image using Trivy
func ScanContainer(ctx context.Context, req *mcp.CallToolRequest, args trivy.ScanParams) (*mcp.CallToolResult, any, error) {
	// Input validation
	if args.Image == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "image parameter is required"}},
		}, nil, fmt.Errorf("image parameter is required")
	}

	req.Session.Log(ctx, &mcp.LoggingMessageParams{
		Data:   fmt.Sprintf("Starting vulnerability scan for image: %s", args.Image),
		Level:  "info",
		Logger: "trivy",
	})

	// Perform the vulnerability scan
	scanResult, err := trivy.Scan(ctx, req.Session, args)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Vulnerability scan failed: %v", err)}},
		}, nil, err
	}

	// Format the scan results with clearer workflow guidance
	var resultMsg strings.Builder
	resultMsg.WriteString(fmt.Sprintf("Vulnerability scan completed for image: %s\n", scanResult.Image))
	resultMsg.WriteString(fmt.Sprintf("Total vulnerabilities found: %d\n", scanResult.VulnCount))
	resultMsg.WriteString(fmt.Sprintf("Scanned platforms: %s\n", strings.Join(scanResult.Platforms, ", ")))
	resultMsg.WriteString(fmt.Sprintf("Report directory: %s\n", scanResult.ReportPath))
	resultMsg.WriteString("\n=== NEXT STEPS ===")
	resultMsg.WriteString("\nTo patch vulnerabilities found in this scan, use the 'patch-vulnerabilities' tool with the above report directory path.")
	resultMsg.WriteString("\n\nNOTE: Do NOT use 'patch-platforms' or 'patch-comprehensive' if you want to patch based on these scan results.")
	resultMsg.WriteString("\nThose tools are for patching WITHOUT vulnerability scanning.")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultMsg.String()}},
	}, nil, nil
}

func Version(ctx context.Context, req *mcp.CallToolRequest, args map[string]interface{}) (*mcp.CallToolResult, any, error) {
	cmd := exec.Command("copa", "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	version := string(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: version}},
	}, nil, nil
}

func WorkflowGuide(ctx context.Context, req *mcp.CallToolRequest, args map[string]interface{}) (*mcp.CallToolResult, any, error) {
	guidance := getWorkflowGuidance()
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: guidance}},
	}, nil, nil
}
