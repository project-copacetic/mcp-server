package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Auth interface for registry authentication operations
type Auth interface {
	SetupRegistryAuthFromEnv() (bool, error)
}

// AuthImpl implements Auth using real docker operations
type AuthImpl struct{}

func (d *AuthImpl) SetupRegistryAuthFromEnv() (bool, error) {
	return SetupRegistryAuthFromEnv()
}

// LoginWithToken authenticates to a registry using a token via docker login
func LoginWithToken(registry, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token cannot be empty")
	}

	// Default to Docker Hub if no registry specified
	if registry == "" {
		registry = "docker.io"
	}

	// Use docker login with --password-stdin for security
	cmd := exec.Command("docker", "login", registry, "-u", "_token", "--password-stdin")
	cmd.Stdin = strings.NewReader(token)

	// Capture both stdout and stderr for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("docker login failed: %v\nOutput: %s", err, string(output))
	}

	return true, nil
}

// SetupRegistryAuthFromEnv reads registry token from environment and runs docker login
// Environment variables:
// - REGISTRY_TOKEN: The authentication token
// - REGISTRY_HOST: The registry hostname (optional, defaults to docker.io)
func SetupRegistryAuthFromEnv() (bool, error) {
	token := os.Getenv("REGISTRY_TOKEN")
	if token == "" {
		// Not an error - just means no token authentication requested
		return false, nil
	}

	registry := os.Getenv("REGISTRY_HOST")
	// Leave empty to use Docker's default behavior

	return LoginWithToken(registry, token)
}
