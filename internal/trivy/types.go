package trivy

// ScanResult - result of a vulnerability scan
type ScanResult struct {
	Image         string
	ReportPath    string
	VulnCount     int
	Platforms     []string
	ScanCompleted bool
}

// ScanParams - parameters for scanning container images for vulnerabilities
type ScanParams struct {
	Image    string   `json:"image" jsonschema:"the image reference of the container to scan for vulnerabilities"`
	Platform []string `json:"platform,omitempty" jsonschema:"Target platform(s) for vulnerability scanning (e.g., linux/amd64,linux/arm64). Valid platforms: linux/amd64, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/arm/v7, linux/arm/v6. If not specified, scans the host platform"`
}
