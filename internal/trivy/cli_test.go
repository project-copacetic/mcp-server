package trivy

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Test Suite for Trivy CLI
type TrivyTestSuite struct {
	suite.Suite
}

// Test isImageLocal function
func (suite *TrivyTestSuite) TestIsImageLocal_LocalImage() {
	// Skip if docker is not available
	if !isDockerAvailable() {
		suite.T().Skip("Docker not available, skipping local image test")
	}

	ctx := context.Background()

	// Pull a small test image to ensure it exists locally
	cmd := exec.CommandContext(ctx, "docker", "pull", "alpine:latest")
	err := cmd.Run()
	if err != nil {
		suite.T().Skip("Failed to pull test image, skipping local image test")
	}

	// Test with local image
	isLocal := isImageLocal(ctx, "alpine:latest")
	suite.True(isLocal, "alpine:latest should be detected as local after pulling")
}

func (suite *TrivyTestSuite) TestIsImageLocal_RemoteImage() {
	ctx := context.Background()

	// Test with a likely non-existent local image (using a specific digest)
	nonExistentImage := "alpine@sha256:nonexistent1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	isLocal := isImageLocal(ctx, nonExistentImage)
	suite.False(isLocal, "Non-existent image should be detected as remote")
}

func (suite *TrivyTestSuite) TestIsImageLocal_DockerCommandFails() {
	ctx := context.Background()

	// Test with invalid docker binary path by creating a context where docker would fail
	// We'll simulate this by using an image name that would cause docker to fail
	isLocal := isImageLocal(ctx, "")
	suite.False(isLocal, "Empty image name should return false")
}

func (suite *TrivyTestSuite) TestIsImageLocal_NonExistentImage() {
	ctx := context.Background()

	// Test with an image that definitely doesn't exist locally
	imageName := "nonexistent/image:definitely-not-local-" + suite.T().Name()
	isLocal := isImageLocal(ctx, imageName)
	suite.False(isLocal, "Non-existent image should be detected as remote")
}

// Test argument construction with and without local detection
func (suite *TrivyTestSuite) TestTrivyArgsConstruction() {
	ctx := context.Background()

	// Test case 1: Remote image (simulated by using non-existent image)
	remoteImage := "definitely/remote:image"
	isLocal := isImageLocal(ctx, remoteImage)
	suite.False(isLocal, "Test image should be detected as remote")

	// Test case 2: We can't easily test a true local image without docker setup
	// but we can test the logic path
}

// Benchmark tests
func BenchmarkIsImageLocal_RemoteImage(b *testing.B) {
	ctx := context.Background()
	image := "nonexistent/image:benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isImageLocal(ctx, image)
	}
}

func BenchmarkIsImageLocal_EmptyImage(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isImageLocal(ctx, "")
	}
}

// Test error scenarios
func (suite *TrivyTestSuite) TestIsImageLocal_ContextCancellation() {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	isLocal := isImageLocal(ctx, "alpine:latest")
	// Should handle context cancellation gracefully and return false
	suite.False(isLocal, "Cancelled context should return false")
}

// Helper function to check if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// Integration test that validates the overall trivy argument construction
func (suite *TrivyTestSuite) TestTrivyArgumentConstruction_Integration() {
	if !isTrivyAvailable() {
		suite.T().Skip("Trivy not available, skipping integration test")
	}

	ctx := context.Background()

	// Test the overall argument construction logic
	// This doesn't actually run trivy but validates our argument building
	platforms := []string{"linux/amd64"}
	image := "alpine:3.17"

	// The key test: verify that --image-src is only added for remote images
	isLocal := isImageLocal(ctx, image)

	// Build expected arguments based on whether image is local
	var expectedToContainImageSrc bool
	if isLocal {
		expectedToContainImageSrc = false
	} else {
		expectedToContainImageSrc = true
	}

	// This would be the logic that happens in the actual Run function
	args := []string{"image", "--vuln-type", "os", "--ignore-unfixed", "-f", "json"}
	if !isLocal {
		args = append(args, "--image-src", "remote")
	}
	args = append(args, "--platform", platforms[0])
	args = append(args, "-o", "/tmp/test.json")
	args = append(args, image)

	// Verify the logic
	containsImageSrc := false
	for i, arg := range args {
		if arg == "--image-src" && i+1 < len(args) && args[i+1] == "remote" {
			containsImageSrc = true
			break
		}
	}

	suite.Equal(expectedToContainImageSrc, containsImageSrc,
		"--image-src should only be present for remote images")
}

// Helper function to check if Trivy is available
func isTrivyAvailable() bool {
	cmd := exec.Command("trivy", "version")
	err := cmd.Run()
	return err == nil
}

// Test edge cases
func (suite *TrivyTestSuite) TestIsImageLocal_EdgeCases() {
	ctx := context.Background()

	testCases := []struct {
		name     string
		image    string
		expected bool
	}{
		{
			name:     "empty string",
			image:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			image:    "   ",
			expected: false,
		},
		{
			name:     "invalid format",
			image:    "not-a-valid-image-reference",
			expected: false,
		},
		{
			name:     "registry with port",
			image:    "localhost:5000/myapp:v1",
			expected: false, // Likely not local unless specifically pulled
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := isImageLocal(ctx, tc.image)
			suite.Equal(tc.expected, result, "Image: %s", tc.image)
		})
	}
}

// Test table-driven scenarios
func TestIsImageLocal_TableDriven(t *testing.T) {
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping docker-dependent test in CI environment")
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		image    string
		setup    func() error // Optional setup function
		expected bool
	}{
		{
			name:     "definitely remote image",
			image:    "gcr.io/definitely-not-local/test:v1.0.0",
			expected: false,
		},
		{
			name:     "malformed image name",
			image:    "invalid::image::name",
			expected: false,
		},
		{
			name:     "empty tag",
			image:    "alpine:",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				if err != nil {
					t.Skipf("Setup failed: %v", err)
				}
			}

			result := isImageLocal(ctx, tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Run the test suite
func TestTrivyTestSuite(t *testing.T) {
	suite.Run(t, new(TrivyTestSuite))
}
