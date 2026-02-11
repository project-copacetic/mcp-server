package copa

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/project-copacetic/mcp-server/internal/docker"
	"github.com/project-copacetic/mcp-server/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock for docker auth package
type MockDockerAuth struct {
	mock.Mock
}

func (m *MockDockerAuth) SetupRegistryAuthFromEnv() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// Ensure MockDockerAuth implements docker.Auth interface
var _ docker.Auth = (*MockDockerAuth)(nil)

// Test Suite for CLI
type CLITestSuite struct {
	suite.Suite
	cli *CLI
}

func (suite *CLITestSuite) SetupTest() {
	// Create a basic CLI instance for each test
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "patched",
		Push:  false,
	}
	suite.cli = New(params, false)
}

// Test New function with different parameter types
func (suite *CLITestSuite) TestNew_ComprehensivePatchParams() {
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "test-tag",
		Push:  true,
	}

	cli := New(params, true)

	suite.Equal("alpine:3.17", cli.image)
	suite.Equal("test-tag", cli.tag)
	suite.True(cli.push)
	suite.True(cli.dryRun)
	suite.Equal("copa", cli.copaPath)
	suite.Empty(cli.platforms)
	suite.Empty(cli.reportPath)
}

func (suite *CLITestSuite) TestNew_PlatformSelectivePatchParams() {
	params := types.PlatformSelectivePatchParams{
		Image:    "nginx:latest",
		Tag:      "selective-patch",
		Push:     false,
		Platform: []string{"linux/amd64", "linux/arm64"},
	}

	cli := New(params, false)

	suite.Equal("nginx:latest", cli.image)
	suite.Equal("selective-patch", cli.tag)
	suite.False(cli.push)
	suite.False(cli.dryRun)
	suite.Equal([]string{"linux/amd64", "linux/arm64"}, cli.platforms)
	suite.Empty(cli.reportPath)
}

func (suite *CLITestSuite) TestNew_ReportBasedPatchParams() {
	params := types.ReportBasedPatchParams{
		Image:      "redis:6",
		Tag:        "report-patch",
		Push:       true,
		ReportPath: "/tmp/reports",
	}

	cli := New(params, false)

	suite.Equal("redis:6", cli.image)
	suite.Equal("report-patch", cli.tag)
	suite.True(cli.push)
	suite.Equal("/tmp/reports", cli.reportPath)
	suite.Empty(cli.platforms)
}

// Test Build method
func (suite *CLITestSuite) TestBuild() {
	result := suite.cli.Build()

	// Should return self for chaining
	suite.Equal(suite.cli, result)

	// Check command construction
	suite.NotNil(suite.cli.cmd)
	expectedArgs := []string{"patch", "--image", "alpine:3.17", "--tag", "patched"}
	suite.Equal(expectedArgs, suite.cli.cmd.Args[1:]) // Skip the program name
}

func (suite *CLITestSuite) TestBuild_WithPush() {
	suite.cli.push = true
	suite.cli.Build()

	expectedArgs := []string{"patch", "--image", "alpine:3.17", "--tag", "patched", "--push"}
	suite.Equal(expectedArgs, suite.cli.cmd.Args[1:])
}

func (suite *CLITestSuite) TestBuild_WithoutTag() {
	suite.cli.tag = ""
	suite.cli.Build()

	expectedArgs := []string{"patch", "--image", "alpine:3.17"}
	suite.Equal(expectedArgs, suite.cli.cmd.Args[1:])
}

// Test BuildWithPlatforms method
func (suite *CLITestSuite) TestBuildWithPlatforms() {
	suite.cli.platforms = []string{"linux/amd64", "linux/arm64"}

	result := suite.cli.BuildWithPlatforms()

	suite.Equal(suite.cli, result)
	suite.Contains(suite.cli.cmd.Args, "--platform")
	suite.Contains(suite.cli.cmd.Args, "linux/amd64,linux/arm64")
}

func (suite *CLITestSuite) TestBuildWithPlatforms_UnsupportedPlatforms() {
	// Mix of supported and unsupported platforms
	suite.cli.platforms = []string{"linux/amd64", "windows/amd64", "linux/arm64"}

	suite.cli.BuildWithPlatforms()

	// Should only include supported platforms
	suite.Contains(suite.cli.cmd.Args, "--platform")
	suite.Contains(suite.cli.cmd.Args, "linux/amd64,linux/arm64")
	suite.NotContains(suite.cli.cmd.Args, "windows/amd64")
}

func (suite *CLITestSuite) TestBuildWithPlatforms_EmptyPlatforms() {
	suite.cli.platforms = []string{}

	suite.cli.BuildWithPlatforms()

	// Should not add platform argument when empty
	suite.NotContains(suite.cli.cmd.Args, "--platform")
}

// Test BuildWithReport method
func (suite *CLITestSuite) TestBuildWithReport() {
	suite.cli.reportPath = "/tmp/test-report"

	result := suite.cli.BuildWithReport()

	suite.Equal(suite.cli, result)
	suite.Contains(suite.cli.cmd.Args, "--report")
	suite.Contains(suite.cli.cmd.Args, "/tmp/test-report")
	suite.Contains(suite.cli.cmd.Args, "--output")
	suite.NotEmpty(suite.cli.vexPath)
}

func (suite *CLITestSuite) TestBuildWithReport_EmptyReportPath() {
	suite.cli.reportPath = ""

	suite.cli.BuildWithReport()

	// Should not add report arguments when reportPath is empty
	suite.NotContains(suite.cli.cmd.Args, "--report")
	suite.NotContains(suite.cli.cmd.Args, "--output")
}

// Test validateCommand method
func (suite *CLITestSuite) TestValidateCommand_Success() {
	suite.cli.Build()

	err := suite.cli.validateCommand()

	suite.NoError(err)
}

func (suite *CLITestSuite) TestValidateCommand_NoCommandBuilt() {
	// Don't call Build()

	err := suite.cli.validateCommand()

	suite.Error(err)
	suite.Contains(err.Error(), "no command built")
}

func (suite *CLITestSuite) TestValidateCommand_EmptyImage() {
	suite.cli.image = ""
	suite.cli.Build()

	err := suite.cli.validateCommand()

	suite.Error(err)
	suite.Contains(err.Error(), "image is required")
}

func (suite *CLITestSuite) TestValidateCommand_UnsupportedPlatforms() {
	suite.cli.platforms = []string{"windows/amd64", "darwin/amd64"}
	suite.cli.Build()

	err := suite.cli.validateCommand()

	suite.Error(err)
	suite.Contains(err.Error(), "no supported platforms found")
}

func (suite *CLITestSuite) TestValidateCommand_ReportPathNotExists() {
	suite.cli.reportPath = "/nonexistent/path"
	suite.cli.Build()

	err := suite.cli.validateCommand()

	suite.Error(err)
	suite.Contains(err.Error(), "report path does not exist")
}

func (suite *CLITestSuite) TestValidateCommand_ReportPathExists() {
	// Create a temporary file to test existing path
	tmpFile, err := os.CreateTemp("", "test-report-*")
	suite.Require().NoError(err)
	defer os.Remove(tmpFile.Name())

	suite.cli.reportPath = tmpFile.Name()
	suite.cli.Build()

	err = suite.cli.validateCommand()

	suite.NoError(err)
}

// Test execute method with dry run
func (suite *CLITestSuite) TestExecute_DryRun() {
	suite.cli.dryRun = true
	suite.cli.Build()

	ctx := context.Background()
	result, err := suite.cli.execute(ctx)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Greater(result.Duration, time.Duration(0))
	suite.Empty(result.Output)
	suite.Empty(result.Error)
	suite.Equal(0, result.ExitCode)
}

func (suite *CLITestSuite) TestExecute_NoCommandBuilt() {
	ctx := context.Background()

	result, err := suite.cli.execute(ctx)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "no command built")
}

// Test Run method integration
func (suite *CLITestSuite) TestRun_ValidationFails() {
	// Don't build command to trigger validation failure
	ctx := context.Background()

	result, err := suite.cli.Run(ctx)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "command validation failed")
}

func (suite *CLITestSuite) TestRun_WithMockDockerAuth_PushFlagAdded() {
	// Create mock docker auth
	mockAuth := new(MockDockerAuth)
	mockAuth.On("SetupRegistryAuthFromEnv").Return(true, nil)

	// Create CLI with mock docker auth
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "test-patched",
		Push:  false, // Initially false
	}
	cli := NewWithDockerAuth(params, true, mockAuth) // Use dry run to avoid actual execution

	// Build the command first
	cli.Build()

	// Verify --push is not initially in the command
	suite.NotContains(cli.cmd.Args, "--push")

	ctx := context.Background()
	result, err := cli.Run(ctx) // Use the cli instance with mock, not suite.cli

	// Verify the mock was called
	mockAuth.AssertExpectations(suite.T())

	// Should succeed since it's a dry run
	suite.NoError(err)
	suite.NotNil(result)

	// Verify that push was set to true
	suite.True(cli.push)

	// Verify --push flag was added to command arguments
	suite.Contains(cli.cmd.Args, "--push")
}

func (suite *CLITestSuite) TestRun_WithMockDockerAuth_NoPushWhenAuthFalse() {
	// Create mock docker auth that returns false
	mockAuth := new(MockDockerAuth)
	mockAuth.On("SetupRegistryAuthFromEnv").Return(false, nil)

	// Create CLI with mock docker auth
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "test-patched",
		Push:  false,
	}
	cli := NewWithDockerAuth(params, true, mockAuth) // Use dry run

	// Build the command first
	cli.Build()

	// Verify --push is not initially in the command
	suite.NotContains(cli.cmd.Args, "--push")

	ctx := context.Background()
	result, err := cli.Run(ctx) // Use the cli instance with mock

	// Verify the mock was called
	mockAuth.AssertExpectations(suite.T())

	// Should succeed since it's a dry run
	suite.NoError(err)
	suite.NotNil(result)

	// Verify that push remains false
	suite.False(cli.push)

	// Verify --push flag was NOT added to command arguments
	suite.NotContains(cli.cmd.Args, "--push")
}

func (suite *CLITestSuite) TestRun_WithMockDockerAuth_AuthError() {
	// Create mock docker auth that returns an error
	mockAuth := new(MockDockerAuth)
	expectedError := fmt.Errorf("authentication failed")
	mockAuth.On("SetupRegistryAuthFromEnv").Return(false, expectedError)

	// Create CLI with mock docker auth
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "test-patched",
		Push:  false,
	}
	cli := NewWithDockerAuth(params, false, mockAuth)

	// Build the command first
	cli.Build()

	ctx := context.Background()
	result, err := cli.Run(ctx) // Use the cli instance with mock

	// Verify the mock was called
	mockAuth.AssertExpectations(suite.T())

	// Should fail due to auth error
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "authentication setup failed")
	suite.Contains(err.Error(), "authentication failed")
}

// Test platform support functions
func TestIsPlatformSupported(t *testing.T) {
	tests := []struct {
		platform string
		expected bool
	}{
		{"linux/amd64", true},
		{"linux/arm64", true},
		{"linux/arm64/v8", true}, // Should be supported as arm64 variant
		{"linux/arm/v7", true},
		{"linux/386", true},
		{"windows/amd64", false},
		{"darwin/amd64", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			result := IsPlatformSupported(tt.platform)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterSupportedPlatforms(t *testing.T) {
	tests := []struct {
		name      string
		platforms []string
		expected  []string
	}{
		{
			name:      "all supported",
			platforms: []string{"linux/amd64", "linux/arm64"},
			expected:  []string{"linux/amd64", "linux/arm64"},
		},
		{
			name:      "mixed supported and unsupported",
			platforms: []string{"linux/amd64", "windows/amd64", "linux/arm64"},
			expected:  []string{"linux/amd64", "linux/arm64"},
		},
		{
			name:      "none supported",
			platforms: []string{"windows/amd64", "darwin/amd64"},
			expected:  nil,
		},
		{
			name:      "empty input",
			platforms: []string{},
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterSupportedPlatforms(tt.platforms)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test ExecutionResult struct
func TestExecutionResult(t *testing.T) {
	result := &ExecutionResult{
		ExitCode: 0,
		Output:   "success output",
		Error:    "some warning",
		Duration: time.Second * 5,
		VexPath:  "/tmp/vex.json",
	}

	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "success output", result.Output)
	assert.Equal(t, "some warning", result.Error)
	assert.Equal(t, time.Second*5, result.Duration)
	assert.Equal(t, "/tmp/vex.json", result.VexPath)
}

// Test error scenarios
func (suite *CLITestSuite) TestExecute_CommandFailure() {
	// Use 'false' command which exists on all Unix systems and always fails with exit code 1
	// This doesn't depend on copa being installed but tests command failure scenarios
	suite.cli.copaPath = "false"
	suite.cli.Build()

	ctx := context.Background()
	result, err := suite.cli.execute(ctx)

	// Should return an error but still provide result with details
	suite.Error(err)
	suite.NotNil(result)
	suite.Contains(err.Error(), "command execution failed")
	suite.Greater(result.Duration, time.Duration(0))
	suite.Equal(1, result.ExitCode) // false command always exits with code 1
}

// Benchmark tests
func BenchmarkCLIBuild(b *testing.B) {
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "bench-test",
		Push:  false,
	}

	for i := 0; i < b.N; i++ {
		cli := New(params, false)
		cli.Build()
	}
}

func BenchmarkCLIValidation(b *testing.B) {
	params := types.ComprehensivePatchParams{
		Image: "alpine:3.17",
		Tag:   "bench-test",
		Push:  false,
	}
	cli := New(params, false)
	cli.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cli.validateCommand()
	}
}

// Run the test suite
func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

// Table-driven tests for comprehensive parameter testing
func TestNew_AllParameterTypes(t *testing.T) {
	tests := []struct {
		name     string
		params   interface{}
		expected struct {
			image      string
			tag        string
			push       bool
			platforms  []string
			reportPath string
		}
	}{
		{
			name: "ComprehensivePatchParams",
			params: types.ComprehensivePatchParams{
				Image: "test:1.0",
				Tag:   "comp-patch",
				Push:  true,
			},
			expected: struct {
				image      string
				tag        string
				push       bool
				platforms  []string
				reportPath string
			}{
				image:      "test:1.0",
				tag:        "comp-patch",
				push:       true,
				platforms:  nil,
				reportPath: "",
			},
		},
		{
			name: "PlatformSelectivePatchParams",
			params: types.PlatformSelectivePatchParams{
				Image:    "test:2.0",
				Tag:      "plat-patch",
				Push:     false,
				Platform: []string{"linux/amd64"},
			},
			expected: struct {
				image      string
				tag        string
				push       bool
				platforms  []string
				reportPath string
			}{
				image:      "test:2.0",
				tag:        "plat-patch",
				push:       false,
				platforms:  []string{"linux/amd64"},
				reportPath: "",
			},
		},
		{
			name: "ReportBasedPatchParams",
			params: types.ReportBasedPatchParams{
				Image:      "test:3.0",
				Tag:        "report-patch",
				Push:       true,
				ReportPath: "/tmp/reports",
			},
			expected: struct {
				image      string
				tag        string
				push       bool
				platforms  []string
				reportPath string
			}{
				image:      "test:3.0",
				tag:        "report-patch",
				push:       true,
				platforms:  nil,
				reportPath: "/tmp/reports",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli *CLI

			switch p := tt.params.(type) {
			case types.ComprehensivePatchParams:
				cli = New(p, false)
			case types.PlatformSelectivePatchParams:
				cli = New(p, false)
			case types.ReportBasedPatchParams:
				cli = New(p, false)
			}

			assert.Equal(t, tt.expected.image, cli.image)
			assert.Equal(t, tt.expected.tag, cli.tag)
			assert.Equal(t, tt.expected.push, cli.push)
			assert.Equal(t, tt.expected.platforms, cli.platforms)
			assert.Equal(t, tt.expected.reportPath, cli.reportPath)
		})
	}
}

// Test context cancellation
func (suite *CLITestSuite) TestExecute_ContextCancellation() {
	// Use 'sleep' command to test context cancellation without depending on copa
	suite.cli.copaPath = "sleep"
	suite.cli.Build()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := suite.cli.execute(ctx)

	// Should handle context cancellation gracefully
	if err != nil {
		// Context cancellation should be in the error
		suite.Contains(err.Error(), "context")
	}
	suite.NotNil(result)
}
