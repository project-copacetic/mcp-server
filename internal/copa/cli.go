package copa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/openvex/go-vex/pkg/vex"
	"github.com/project-copacetic/mcp-server/internal/docker"
	"github.com/project-copacetic/mcp-server/internal/types"
)

const (
	defaultVexFile = "vex.json"
)

// TODO: improve error handling
// ExecutionResult holds the result of command execution
type ExecutionResult struct {
	ExitCode                int
	Output                  string
	Error                   string
	Duration                time.Duration
	VexPath                 string // Only populated for report-based patching
	UpdatedPackageCount     int
	FixedVulnerabilityCount int
}

// CopaSupportedPlatforms lists all platforms that Copa can patch
// Based on Copa documentation: https://project-copacetic.github.io/copacetic/website/multiplatform-patching
var CopaSupportedPlatforms = []string{
	"linux/amd64",
	"linux/arm64",
	"linux/arm/v7",
	"linux/arm/v6",
	"linux/386",
	"linux/ppc64le",
	"linux/s390x",
	"linux/riscv64",
}

type CLI struct {
	copaPath   string
	dryRun     bool
	image      string
	tag        string
	platforms  []string
	push       bool
	reportPath string
	vexPath    string
	cmd        *exec.Cmd   // Current command being built
	dockerAuth docker.Auth // Dependency injection for docker authentication
}

type PatchParamsConstraint interface {
	types.ReportBasedPatchParams |
		types.PlatformSelectivePatchParams |
		types.ComprehensivePatchParams
}

// NOTE: use generic for param types to assist the agent with populating the correct values.
func New[T PatchParamsConstraint](params T, dryRun bool) *CLI {
	var image, tag, reportPath string
	var platforms []string
	var push bool

	// Extract common fields using type switch
	switch p := any(params).(type) {
	case types.ReportBasedPatchParams:
		image, tag, push, reportPath = p.Image, p.Tag, p.Push, p.ReportPath
	case types.PlatformSelectivePatchParams:
		image, tag, push, platforms = p.Image, p.Tag, p.Push, p.Platform
	case types.ComprehensivePatchParams:
		image, tag, push = p.Image, p.Tag, p.Push
	}

	return &CLI{
		copaPath:   "copa",
		dryRun:     dryRun,
		image:      image,
		tag:        tag,
		platforms:  platforms,
		push:       push,
		reportPath: reportPath,
		dockerAuth: &docker.AuthImpl{}, // Default to real implementation
	}
}

// NewWithDockerAuth creates a CLI instance with custom docker auth for testing
func NewWithDockerAuth[T PatchParamsConstraint](params T, dryRun bool, dockerAuth docker.Auth) *CLI {
	cli := New(params, dryRun)
	cli.dockerAuth = dockerAuth
	return cli
}

func (c *CLI) Build() *CLI {
	args := []string{"patch"}
	args = append(args, "--image", c.image)

	if c.tag != "" {
		args = append(args, "--tag", c.tag)
	}

	if c.push {
		args = append(args, "--push")
	}

	c.cmd = exec.Command(c.copaPath, args...)
	return c
}

func (c *CLI) BuildWithPlatforms() *CLI {
	c = c.Build()

	if len(c.platforms) > 0 {
		supportedPlatforms := FilterSupportedPlatforms(c.platforms)
		if len(supportedPlatforms) > 0 {
			c.cmd.Args = append(c.cmd.Args, "--platform", strings.Join(supportedPlatforms, ","))
		}
	}

	return c
}

func (c *CLI) BuildWithReport() *CLI {
	c = c.Build()

	if c.reportPath != "" {
		c.cmd.Args = append(c.cmd.Args, "--report", c.reportPath)
	}

	return c
}

func (c *CLI) setupAuth() error {
	// Check if we need remote patching (push to registry)
	remotePatch, err := c.dockerAuth.SetupRegistryAuthFromEnv()
	if err != nil {
		return fmt.Errorf("failed to authenticate to registry: %w", err)
	}

	if remotePatch && c.cmd != nil && !slices.Contains(c.cmd.Args, "--push") {
		c.push = true
		c.cmd.Args = append(c.cmd.Args, "--push")
	}

	return nil
}

func (c *CLI) setupVexDir() error {
	if c.reportPath != "" {
		path, err := os.MkdirTemp(os.TempDir(), "vex-*")
		if err != nil {
			return err
		}
		c.vexPath = filepath.Join(path, defaultVexFile)
		c.cmd.Args = append(c.cmd.Args, "--output", c.vexPath)
	}
	return nil
}

func (c *CLI) validateCommand() error {
	if c.cmd == nil {
		return fmt.Errorf("no command built - call a Build method first")
	}

	if c.image == "" {
		return fmt.Errorf("image is required")
	}

	// Validate platforms if specified
	if len(c.platforms) > 0 {
		supportedPlatforms := FilterSupportedPlatforms(c.platforms)
		if len(supportedPlatforms) == 0 {
			return fmt.Errorf("no supported platforms found in: %v", c.platforms)
		}
		if len(supportedPlatforms) != len(c.platforms) {
			fmt.Fprintf(os.Stderr, "Warning: some platforms not supported by Copa, using: %v\n", supportedPlatforms)
		}
	}

	// Validate report path if specified
	if c.reportPath != "" {
		if _, err := os.Stat(c.reportPath); os.IsNotExist(err) {
			return fmt.Errorf("report path does not exist: %s", c.reportPath)
		}
	}

	return nil
}

func (c *CLI) execute(ctx context.Context) (*ExecutionResult, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("no command built - call a Build method first")
	}

	startTime := time.Now()
	result := &ExecutionResult{}

	if c.dryRun {
		fmt.Fprintf(os.Stderr, "[DRY RUN] %s %s\n", c.cmd.Path, strings.Join(c.cmd.Args[1:], " "))
		result.Duration = time.Since(startTime)
		return result, nil
	}

	var stdout, stderr bytes.Buffer
	c.cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	c.cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	c.cmd = exec.CommandContext(ctx, c.cmd.Path, c.cmd.Args[1:]...)

	fmt.Fprintf(os.Stderr, "Executing: %s %s\n", c.cmd.Path, strings.Join(c.cmd.Args[1:], " "))

	err := c.cmd.Run()

	result.Duration = time.Since(startTime)
	result.Output = stdout.String()
	result.Error = stderr.String()
	result.VexPath = c.vexPath

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}

func (c *CLI) Run(ctx context.Context) (*ExecutionResult, error) {
	if err := c.validateCommand(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	if err := c.setupAuth(); err != nil {
		return nil, fmt.Errorf("authentication setup failed: %w", err)
	}

	if err := c.setupVexDir(); err != nil {
		return nil, fmt.Errorf("creating vex temp dir failed: %w", err)
	}

	result, err := c.execute(ctx)
	if err != nil {
		return result, fmt.Errorf("execution failed: %w", err)
	}

	result.FixedVulnerabilityCount, result.UpdatedPackageCount, err = c.parseVexDoc(c.vexPath)
	if err != nil {
		return result, fmt.Errorf("parsing vex doc failed: %w", err)
	}

	return result, nil
}

// IsPlatformSupported checks if the given platform is supported by Copa for patching
func IsPlatformSupported(platform string) bool {
	for _, supported := range CopaSupportedPlatforms {
		if platform == supported {
			return true
		}
		// Handle arm64 variants - Copa supports "linux/arm64" which covers "linux/arm64/v8"
		if supported == "linux/arm64" && (platform == "linux/arm64/v8" || platform == "linux/arm64") {
			return true
		}
	}
	return false
}

// FilterSupportedPlatforms returns only the platforms that Copa can patch from the given list
func FilterSupportedPlatforms(platforms []string) []string {
	var supported []string
	for _, platform := range platforms {
		if IsPlatformSupported(platform) {
			supported = append(supported, platform)
		}
	}
	return supported
}

func (c *CLI) parseVexDoc(path string) (numFixedVulns, updatedPackageCount int, err error) {
	if c.vexPath == "" {
		return 0, 0, nil
	}

	vexData, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	var doc vex.VEX

	if err := json.Unmarshal(vexData, &doc); err != nil {
		return 0, 0, err
	}

	for _, stmt := range doc.Statements {
		if stmt.Status == vex.StatusFixed {
			numFixedVulns++
			for _, product := range stmt.Products {
				updatedPackageCount += len(product.Subcomponents)
			}
		}
	}
	return numFixedVulns, updatedPackageCount, nil
}
