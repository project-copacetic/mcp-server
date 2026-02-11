FROM golang:1.24.6-alpine AS builder
WORKDIR /app

# Check if binary already exists (GoReleaser context)
COPY . .
RUN if [ ! -f "./copacetic-mcp-server" ]; then \
      echo "Building from source..."; \
      go mod download; \
      CGO_ENABLED=0 GOOS=linux go build -o copacetic-mcp-server .; \
    else \
      echo "Using pre-built binary"; \
    fi

FROM debian:12-slim
WORKDIR /app

# Install dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends tar ca-certificates gnupg curl jq qemu-user-static binfmt-support && \
    install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable" > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends docker-ce docker-ce-cli docker-buildx-plugin containerd.io && \
    rm -rf /var/lib/apt/lists/*

# Install Trivy
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin && trivy --version

# Install Copa
RUN curl --retry 5 -fsSL -o copa.tar.gz https://github.com/project-copacetic/copacetic/releases/download/v0.11.1/copa_0.11.1_linux_amd64.tar.gz && \
    tar -zxvf copa.tar.gz && mv copa /usr/local/bin/ && rm -f copa.tar.gz

# Copy the binary (works for both GoReleaser and standalone builds)
COPY --from=builder /app/copacetic-mcp-server ./copacetic-mcp-server

ENTRYPOINT ["/app/copacetic-mcp-server"]
