package types

type Ver struct {
	Version string `json:"version" jsonschema:"the version of the copa cli"`
}

type PatchResult struct {
	OriginalImage       string
	PatchedImage        []string
	ReportPath          string
	VexPath             string
	NumFixedVulns       int
	UpdatedPackageCount int
	ScanPerformed       bool
	VexGenerated        bool
}

// ReportBasedPatchParams - patches only vulnerabilities found in an existing vulnerability report
// NOTE: This requires a vulnerability scan to be run first using the 'scan-container' tool
type ReportBasedPatchParams struct {
	Image      string `json:"image" jsonschema:"the image reference of the container being patched"`
	Tag        string `json:"patchtag" jsonschema:"the new tag name (not full image reference) for the patched image. Example: 'patched' or 'v1.0-secure', not 'alpine:patched'"`
	Push       bool   `json:"push" jsonschema:"push patched image to destination registry"`
	ReportPath string `json:"reportPath" jsonschema:"Path to the vulnerability report directory created by the 'scan-container' tool. This must be provided - run 'scan-container' first to generate the report."`
}

// PlatformSelectivePatchParams - patches only specified platforms
type PlatformSelectivePatchParams struct {
	Image    string   `json:"image" jsonschema:"the image reference of the container being patched"`
	Tag      string   `json:"patchtag" jsonschema:"the new tag name (not full image reference) for the patched image. Example: 'patched' or 'v1.0-secure', not 'alpine:patched'"`
	Push     bool     `json:"push" jsonschema:"push patched image to destination registry"`
	Platform []string `json:"platform" jsonschema:"Target platform(s) for patching (e.g., linux/amd64,linux/arm64). Valid platforms: linux/amd64, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/arm/v7, linux/arm/v6. Only specified platforms will be patched, others will be preserved unchanged"`
}

// ComprehensivePatchParams - patches all available platforms with latest updates
type ComprehensivePatchParams struct {
	Image string `json:"image" jsonschema:"the image reference of the container being patched"`
	Tag   string `json:"patchtag" jsonschema:"the new tag name (not full image reference) for the patched image. Example: 'patched' or 'v1.0-secure', not 'alpine:patched'"`
	Push  bool   `json:"push" jsonschema:"push patched image to destination registry"`
}
