# copacetic-mcp

A Model Context Protocol (MCP) server for automated container patching using [Copacetic](https://github.com/project-copacetic/copacetic).

## MCP Tools

This server provides the following Model Context Protocol (MCP) tools:

- **`version`**: Get the version of the Copa CLI tool
- **`workflow-guide`**: Get guidance on which Copacetic tools to use for different container patching scenarios
- **`scan-container`**: Scan container images for vulnerabilities using Trivy - creates vulnerability reports required for report-based patching
- **`patch-report-based`**: Patch container image vulnerabilities using a pre-generated vulnerability report from 'scan-container' tool (RECOMMENDED approach for vulnerability-based patching)
- **`patch-platform-selective`**: Patch specific container image platforms with Copa - patches only the specified platforms WITHOUT vulnerability scanning
- **`patch-comprehensive`**: Comprehensively patch all container image platforms with Copa - patches all available platforms WITHOUT vulnerability scanning

## Installation

### VSCode Setup

To use copacetic-mcp with VSCode, add the MCP server configuration to your `mcp.json` file.

**Opening the MCP Configuration File**:

1. Open the Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`)
2. Choose one of these commands:
   - `MCP: Open User Configuration` - Configure for your user profile (applies to all workspaces)
   - `MCP: Open Workspace Folder Configuration` - Configure for the current workspace only

This will create or open your `mcp.json` file where you can add the following configuration:

```json
{
  "servers": {
    "copaceticMcp": {
      "type": "stdio",
      "command": "/path/to/copacetic-mcp-server",
      "args": []
    }
  }
}
```

Replace `/path/to/copacetic-mcp-server` with the actual path to your copacetic-mcp server binary. You can download the latest release from the [Releases]() page.

> **Note**: When using the binary directly, ensure that the Copacetic CLI, Trivy, and Docker with emulation are installed on your system.

### Docker Container option

You can also run copacetic-mcp using Docker by adding this configuration to your `mcp.json`:

```json
{
  "servers": {
    "copaceticMcpDocker": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--mount",
        "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock",
        "--mount",
        "type=bind,source=${env:HOME}/.docker/config.json,target=/root/.docker/config.json",
        "ghcr.io/project-copacetic/mcp-server:latest"
      ],
      "env": {
        "DOCKER_HOST": "unix:///var/run/docker.sock"
      }
    }
  }
}
```

Notes:

- Mounting the Docker socket gives the container access to the host Docker daemon; this is required for Copacetic image operations but has security implications—only run trusted images.
- Mounting `${HOME}/.docker/config.json` allows the container to use your registry credentials for pulling/pushing images.

<!-- TODO: Docker Gateway / Catalog  -->

#### Alternative: Using with Claude Desktop

You can also configure copacetic-mcp for use with Claude Desktop by adding it to your MCP configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "copacetic-mcp": {
      "command": "/path/to/copacetic-mcp-server",
      "args": []
    }
  }
}
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
