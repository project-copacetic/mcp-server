# Copacetic MCP Server

## Project Overview

Copacetic MCP is a Go application that provides a Model Context Protocol (MCP) server for automated container image patching using Copacetic and Trivy. It exposes container patching capabilities through the MCP protocol, allowing AI agents and tools to patch container image vulnerabilities programmatically.

**Main commands**: MCP tools `version`, `scan-container`, `patch-comprehensive`, `patch-platforms`, `patch-vulnerabilities`, and `workflow-guide`
**Module**: `github.com/project-copacetic/mcp-server`

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Folder Structure

- `cmd/copa-mcp-server/main.go`: Main MCP server entry point
- `cmd/copa-mcp-client/main.go`: CLI client for interacting with MCP server functionality
- `internal/copamcp/`: MCP server setup, tool registration, and protocol handlers
- `internal/copa/`: Copacetic command execution and container patching logic
- `internal/trivy/`: Trivy vulnerability scanning integration
- `internal/types/`: Shared type definitions and execution modes
- `internal/docker/`: Docker authentication utilities
- `.goreleaser.yml`: GoReleaser configuration for cross-platform releases
- `.github/workflows/`: CI/CD automation (build.yml, release.yml)
- `Makefile`: Development tasks and build automation

## Libraries and Frameworks

- **MCP Protocol**: `github.com/modelcontextprotocol/go-sdk/mcp` for Model Context Protocol server implementation
- **CLI Framework**: `github.com/spf13/cobra` for command-line interface structure
- **VEX Support**: `github.com/openvex/go-vex` for vulnerability exchange document generation
- **External Tools**: Copacetic (copa) for patching, Trivy for vulnerability scanning
- **Cross-platform Builds**: GoReleaser for automated multi-platform binary releases

## Coding Standards

- Follow Go best practices and standard formatting (`make fmt`)
- Use `go vet` for static analysis validation (`make vet`)
- Implement comprehensive error handling with wrapped errors using `fmt.Errorf`
- Write tests for new functionality with proper Docker test environment detection
- Use structured logging for debugging and operational visibility
- Follow MCP protocol specifications for tool definitions and responses
- Handle multiplatform scenarios appropriately for container operations

## Key Architecture Concepts

- **MCP Server Architecture**: Provides multiple focused tools through stdin/stdout MCP protocol
- **Tool Workflow**:
  - `scan-container`: Vulnerability scanning with Trivy (creates reports for targeted patching)
  - `patch-vulnerabilities`: Report-based patching (requires scan-container output)
  - `patch-platforms`: Platform-selective patching (no scan, patches specified platforms only)
  - `patch-comprehensive`: Comprehensive patching (no scan, patches all available platforms)
- **Execution Modes**:
  - `report-based`: Patches only vulnerabilities identified through Trivy scanning
  - `platform-selective`: Patches specified platforms without vulnerability scanning
  - `comprehensive`: Patches all available platforms without vulnerability scanning
- **Multiplatform Support**: Handles container images across multiple architectures (amd64, arm64, etc.)
- **External Tool Integration**: Orchestrates Copacetic and Trivy through command execution
- **VEX Generation**: Creates Vulnerability Exchange (VEX) documents for patching results
- **Cross-platform Binary Distribution**: Builds native binaries for Linux, macOS, and Windows

## Supported Container Scenarios

- **Single-arch images**: Direct patching of images for specific platform
- **Multi-arch images**: Platform-specific patching while preserving other architectures
- **Registry Operations**: Pull, patch, and optionally push patched images
- **Tag Management**: Automatic tagging of patched images with `-patched` suffix
- **Vulnerability Reporting**: Integration with Trivy for security scanning

## Key Functions and Components

- `NewServer()`: Creates MCP server instance with registered tools
- `Run()`: Starts MCP server with stdio transport
- `Patch()`: Main patching tool that orchestrates vulnerability scanning and image patching
- `Version()`: Returns Copacetic version information
- `copa.Run()`: Executes Copacetic patching with proper argument construction
- `trivy.Scan()`: Performs vulnerability scanning using Trivy
- `DetermineExecutionMode()`: Selects appropriate patching strategy based on parameters

## Working Effectively

### Prerequisites and Installation

- Go 1.20 or later (tested with Go 1.24.6)
- Docker (for container operations and some tests)
- [Copacetic](https://github.com/project-copacetic/copacetic) v0.8.0+ for container patching
- [Trivy](https://github.com/aquasecurity/trivy) v0.65.0+ for vulnerability scanning
- [GoReleaser](https://goreleaser.com/) v2.5.0+ for releases

### Install Required Dependencies

Install Copacetic:

```bash
wget -O copa.tar.gz https://github.com/project-copacetic/copacetic/releases/download/v0.8.0/copa_0.8.0_linux_amd64.tar.gz
tar -xzf copa.tar.gz
sudo cp copa /usr/local/bin/
copa --version  # Should show: copa version 0.8.0
```

Install Trivy:

```bash
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sudo sh -s -- -b /usr/local/bin
trivy --version  # Should show: Version: 0.65.0
```

Install GoReleaser (for releases):

```bash
curl -sLO https://github.com/goreleaser/goreleaser/releases/download/v2.5.0/goreleaser_Linux_x86_64.tar.gz
tar -xzf goreleaser_Linux_x86_64.tar.gz
sudo cp goreleaser /usr/local/bin/
goreleaser --version  # Should show version 2.5.0
```

### Build and Test Commands

**NEVER CANCEL ANY BUILD OR TEST COMMAND** - All commands may take longer than expected. Always use adequate timeouts.

Build the project:

```bash
make build  # Takes ~40 seconds. NEVER CANCEL. Set timeout to 120+ seconds.
```

Run tests:

```bash
make test  # Takes ~8 seconds. Docker tests are automatically skipped in CI.
```

Format and validate code:

```bash
make fmt    # Takes ~0.2 seconds
make vet    # Takes ~5 seconds
```

Cross-compile for all platforms:

```bash
make cross-compile  # Takes ~1 minute 45 seconds. NEVER CANCEL. Set timeout to 240+ seconds.
```

Build release artifacts:

```bash
make release-snapshot  # Takes ~2 minutes 41 seconds. NEVER CANCEL. Set timeout to 300+ seconds.
```

### Run the Application

Start the MCP server (interactive mode):

```bash
./bin/copacetic-mcp-server stdio
# Server waits for MCP protocol messages on stdin/stdout
# Use Ctrl+C to stop
```

Run the test client (requires server dependencies):

```bash
./bin/copacetic-mcp-client
# Connects to server and tests the 'patch' tool with alpine:3.17
```

## Validation

### Always Run These Steps After Making Changes:

1. **Build validation** - Build succeeds without errors:

   ```bash
   make build  # Set timeout to 120+ seconds, NEVER CANCEL
   ```

2. **Test validation** - All tests pass:

   ```bash
   make test  # Docker tests skip automatically in CI environments
   ```

3. **Code quality validation** - Required for CI to pass:

   ```bash
   make fmt vet  # Both commands must complete successfully
   ```

4. **Integration tests validation** - Test all MCP tools end-to-end:

   ```bash
   make integration-test-quick  # Quick validation (~10 seconds)
   # OR for comprehensive testing:
   make integration-test  # Full validation (~2-5 minutes, requires copa/trivy/docker)
   ```

5. **MCP server functionality validation** - Test server-client communication:
   ```bash
   # Create test script to validate version tool:
   cat > test_mcp.go << 'EOF'
   package main
   import (
       "context"
       "fmt"
       "log"
       "os/exec"
       "github.com/modelcontextprotocol/go-sdk/mcp"
   )
   func main() {
       ctx := context.Background()
       client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
       cmd := exec.Command("./bin/copacetic-mcp-server")
       transport := mcp.NewCommandTransport(cmd)
       session, err := client.Connect(ctx, transport)
       if err != nil { log.Fatal(err) }
       defer session.Close()
       params := &mcp.CallToolParams{Name: "version", Arguments: map[string]any{}}
       res, err := session.CallTool(ctx, params)
       if err != nil { log.Fatalf("CallTool failed: %v", err) }
       if res.IsError { log.Fatal("version tool failed") }
       for _, c := range res.Content {
           fmt.Printf("Success: %s\n", c.(*mcp.TextContent).Text)
       }
   }
   EOF
   go run test_mcp.go  # Should output: Success: copa version 0.8.0
   rm test_mcp.go
   ```

### Cross-Platform Validation

For release builds, validate cross-compilation works:

```bash
make cross-compile  # Set timeout to 240+ seconds, NEVER CANCEL
ls -la bin/  # Should show binaries for linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64.exe
```

### Integration Tests

The project includes comprehensive integration tests under `.scripts/integration-test.sh` that use the copa-mcp-client to test all MCP tools end-to-end:

**Run full integration tests** (requires copa, trivy, and docker):

```bash
make integration-test  # Takes ~2-5 minutes depending on network and image pulls
# OR directly:
./.scripts/integration-test.sh
```

**Run quick integration tests** (only version and list commands):

```bash
make integration-test-quick  # Takes ~10 seconds
# OR directly:
./.scripts/integration-test.sh --quick
```

**Integration test features:**

- Tests all CLI commands: `version`, `list`, `scan-container`, `patch-comprehensive`, `patch-platforms`, `patch-vulnerabilities`
- Validates complete vulnerability-based patching workflow (scan → patch-vulnerabilities)
- Tests error scenarios with invalid inputs
- Uses alpine:3.17 as test image (configurable via TEST_IMAGE environment variable)
- Automatic cleanup of temporary scan reports
- Colored output with detailed success/failure reporting

**Prerequisites for integration tests:**

- All external dependencies installed (copa, trivy, docker)
- Built binaries in `bin/` directory (`make build` first)
- Docker daemon running (for container operations)

The integration tests provide confidence that the MCP server and client work correctly together and that all patching workflows function as expected.

## Important Build and Timing Information

- **Build time**: ~40 seconds (first time with dependencies)
- **Test time**: ~8 seconds (Docker tests automatically skip in CI)
- **Cross-compile time**: ~1 minute 45 seconds
- **Release build time**: ~2 minutes 41 seconds
- **Format/vet time**: <5 seconds combined

**CRITICAL**: NEVER CANCEL long-running commands. Builds and cross-compilation can take several minutes, especially on slower systems. Always set timeouts to at least double the expected time.

## Common Tasks and Troubleshooting

### MCP Server Architecture

The server provides these MCP tools:

- `version`: Returns copa version information
- `scan-container`: Scans container images for vulnerabilities using Trivy
- `patch-comprehensive`: Patches all available platforms without vulnerability scanning
- `patch-platforms`: Patches specific platforms without vulnerability scanning
- `patch-vulnerabilities`: Patches vulnerabilities based on scan results (requires scan-container output)
- `workflow-guide`: Provides guidance on which tools to use for different scenarios

### Dependencies Not Available

If copa or trivy are not installed:

- Tests will still pass (external tool tests are conditional)
- MCP server will fail when tools are called
- Always install dependencies using the exact commands above

### Docker Tests Skipped

Docker tests automatically skip in CI environments (`CI` or `GITHUB_ACTIONS` env vars set). This is expected behavior.

### Build Artifacts

- Binaries: `bin/copacetic-mcp-server`, `bin/copacetic-mcp-client`
- Cross-compiled: `bin/copacetic-mcp-server-{os}-{arch}[.exe]`
- Release artifacts: `dist/` directory (excluded from git)

### Key Project Structure

```
mcp-server/
├── cmd/
│   ├── copa-mcp-server/main.go    # Main MCP server entry point
│   └── copa-mcp-client/main.go    # CLI client for MCP server interaction
├── internal/
│   ├── copamcp/               # MCP server handlers, tool registration, protocol implementation
│   ├── copa/                  # Copacetic command execution and container patching orchestration
│   ├── trivy/                 # Trivy vulnerability scanning integration
│   ├── types/                 # Shared type definitions, execution modes, and parameters
│   └── docker/                # Docker authentication utilities
├── .goreleaser.yml            # GoReleaser configuration for automated releases
├── .github/workflows/         # GitHub Actions CI/CD automation
│   ├── build.yml             # Continuous integration: build, test, lint on every push/PR
│   └── release.yml           # Automated releases with cross-platform binaries on tags
└── Makefile                   # Development tasks: build, test, format, cross-compile, release
```

## CI/CD Integration

- GitHub Actions automatically builds and tests on push/PR
- Release process uses GoReleaser for cross-platform binaries
- Docker tests are automatically skipped in CI environments
- All validation steps (fmt, vet, test, build) must pass for CI success
