package trivy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// isImageLocal checks if an image exists locally in the Docker daemon
func isImageLocal(ctx context.Context, image string) bool {
	// Handle edge cases
	if strings.TrimSpace(image) == "" {
		return false
	}

	// Use docker images command to check if image exists locally
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}", image)
	output, err := cmd.Output()
	if err != nil {
		// If docker command fails, assume remote
		return false
	}

	// If output is not empty, image exists locally
	return strings.TrimSpace(string(output)) != ""
}

func Run(ctx context.Context, cc *mcp.ServerSession, image string, platform []string) (reportPath string, err error) {
	reportPath, err = os.MkdirTemp(os.TempDir(), "reports-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary report directory: %w", err)
	}
	trivyArgs := []string{
		"image",
		"--vuln-type", "os",
		"--ignore-unfixed",
		"-f", "json",
	}

	if len(platform) == 0 {
		trivyArgs = append(trivyArgs, "-o", filepath.Join(reportPath, "report.json"))
		trivyArgs = append(trivyArgs, image)

		trivyCmd := exec.Command("trivy", trivyArgs...)

		cc.Log(ctx, &mcp.LoggingMessageParams{
			Data:   fmt.Sprintf("Executing: %s %s", trivyCmd.Path, strings.Join(trivyCmd.Args[1:], " ")),
			Level:  "info",
			Logger: "trivy",
		})
		var stderrTrivy strings.Builder
		trivyCmd.Stderr = &stderrTrivy

		err = trivyCmd.Run()
		if err != nil {
			exitCode := ""
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = fmt.Sprintf(" (exit code %d)", exitError.ExitCode())
			}
			errorMsg := fmt.Sprintf("trivy command failed%s: %v\n%s", exitCode, err, stderrTrivy.String())
			return "", fmt.Errorf("%s", errorMsg)
		}

		return reportPath, nil
	}

	for _, p := range platform {
		args := trivyArgs

		if !isImageLocal(ctx, image) {
			args = append(args, "--image-src", "remote")
		}

		args = append(args, "--platform", p)
		args = append(args, "-o", filepath.Join(reportPath, strings.ReplaceAll(p, "/", "-")+".json"))
		args = append(args, image)

		trivyCmd := exec.Command("trivy", args...)

		// Log the command being executed using cc.Log to match copa's pattern
		cc.Log(ctx, &mcp.LoggingMessageParams{
			Data:   fmt.Sprintf("Executing: %s %s", trivyCmd.Path, strings.Join(trivyCmd.Args[1:], " ")),
			Level:  "info",
			Logger: "trivy",
		})
		var stderrTrivy strings.Builder
		trivyCmd.Stderr = &stderrTrivy

		err = trivyCmd.Run()
		if err != nil {
			exitCode := ""
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = fmt.Sprintf(" (exit code %d)", exitError.ExitCode())
			}
			errorMsg := fmt.Sprintf("trivy command failed%s: %v\n%s", exitCode, err, stderrTrivy.String())
			return "", fmt.Errorf("%s", errorMsg)
		}
	}

	return reportPath, nil
}

// Scan performs vulnerability scanning and returns detailed scan results
func Scan(ctx context.Context, cc *mcp.ServerSession, params ScanParams) (*ScanResult, error) {
	reportPath, err := Run(ctx, cc, params.Image, params.Platform)
	if err != nil {
		return nil, fmt.Errorf("vulnerability scan failed: %w", err)
	}

	// Count vulnerabilities in the report(s)
	vulnCount, err := countVulnerabilitiesInReport(reportPath)
	if err != nil {
		cc.Log(ctx, &mcp.LoggingMessageParams{
			Data:   fmt.Sprintf("Warning: Could not count vulnerabilities in report: %v", err),
			Level:  "warn",
			Logger: "trivy",
		})
		vulnCount = 0
	}

	platforms := params.Platform
	if len(platforms) == 0 {
		platforms = []string{"host platform"}
	}

	return &ScanResult{
		Image:         params.Image,
		ReportPath:    reportPath,
		VulnCount:     vulnCount,
		Platforms:     platforms,
		ScanCompleted: true,
	}, nil
}

// countVulnerabilitiesInReport counts total vulnerabilities across all report files
func countVulnerabilitiesInReport(reportPath string) (int, error) {
	// Read directory to find all JSON report files
	entries, err := os.ReadDir(reportPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read report directory: %w", err)
	}

	totalVulns := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(reportPath, entry.Name())
			vulns, err := countVulnerabilitiesInFile(filePath)
			if err != nil {
				return 0, fmt.Errorf("failed to count vulnerabilities in %s: %w", filePath, err)
			}
			totalVulns += vulns
		}
	}

	return totalVulns, nil
}

// countVulnerabilitiesInFile counts vulnerabilities in a single JSON report file
func countVulnerabilitiesInFile(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read report file: %w", err)
	}

	var report struct {
		Results []struct {
			Vulnerabilities []interface{} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(data, &report); err != nil {
		return 0, fmt.Errorf("failed to parse JSON report: %w", err)
	}

	totalVulns := 0
	for _, result := range report.Results {
		totalVulns += len(result.Vulnerabilities)
	}

	return totalVulns, nil
}
