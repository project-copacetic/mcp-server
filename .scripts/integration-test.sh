#!/bin/bash

# Integration test script for Copacetic MCP CLI
# Tests all CLI commands to ensure they function correctly

set -e  # Exit on any error

# Configuration
TEST_IMAGE="alpine:3.17"
CLIENT_BINARY="./bin/copacetic-mcp-client"
REPORT_DIR=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if client binary exists
    if [[ ! -f "$CLIENT_BINARY" ]]; then
        error "Client binary not found at $CLIENT_BINARY"
        log "Run 'make build' first"
        exit 1
    fi
    
    # Check external dependencies
    deps=("copa" "trivy" "docker")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            error "$dep is not installed or not in PATH"
            warn "Integration tests require all dependencies to be installed"
            exit 1
        fi
    done
    
    log "All prerequisites satisfied"
}

# Test individual CLI commands
test_version_command() {
    log "Testing version command..."
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY version 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        if echo "$output" | grep -q "copa"; then
            success "✓ Version command succeeded"
            echo "  Version: $(echo "$output" | grep copa)"
            return 0
        else
            error "✗ Version command output doesn't contain 'copa'"
            echo "$output"
            return 1
        fi
    else
        error "✗ Version command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

test_list_command() {
    log "Testing list command..."
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY list 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        local expected_tools=("version" "scan-container" "patch-comprehensive" "patch-platform-selective" "patch-report-based")
        local success=true
        
        for tool in "${expected_tools[@]}"; do
            if echo "$output" | grep -q "$tool"; then
                log "  ✓ Found tool: $tool"
            else
                error "  ✗ Missing tool: $tool"
                success=false
            fi
        done
        
        if $success; then
            success "✓ List command succeeded"
            return 0
        else
            error "✗ List command failed - missing tools"
            echo "$output"
            return 1
        fi
    else
        error "✗ List command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

test_scan_command() {
    log "Testing scan command..."
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY scan-container --image "$TEST_IMAGE" 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        if echo "$output" | grep -q "Vulnerability scan completed"; then
            success "✓ Scan command succeeded"
            
            # Extract report directory for later use
            REPORT_DIR=$(echo "$output" | grep "Report directory:" | sed 's/.*Report directory: //' | tr -d '\n\r')
            if [[ -n "$REPORT_DIR" ]]; then
                log "  Report directory: $REPORT_DIR"
            else
                warn "  Could not extract report directory"
            fi
            return 0
        else
            error "✗ Scan command output doesn't indicate success"
            echo "$output"
            return 1
        fi
    else
        error "✗ Scan command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

test_patch_vulnerabilities_command() {
    log "Testing patch-vulnerabilities command..."
    
    if [[ -z "$REPORT_DIR" ]]; then
        error "No report directory available for vulnerability patching"
        log "Skipping patch-vulnerabilities test - requires successful scan first"
        return 1
    fi
    
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY patch-vulnerabilities \
        --image "$TEST_IMAGE" \
        --patchtag "integration-test-vuln" \
        --report-path "$REPORT_DIR" 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        if echo "$output" | grep -q "successful patched"; then
            success "✓ Patch vulnerabilities command succeeded"
            echo "  $(echo "$output" | grep "successful patched")"
            return 0
        else
            error "✗ Patch vulnerabilities command output doesn't indicate success"
            echo "$output"
            return 1
        fi
    else
        error "✗ Patch vulnerabilities command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

test_patch_comprehensive_command() {
    log "Testing patch-comprehensive command..."
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY patch-comprehensive \
        --image "$TEST_IMAGE" \
        --patchtag "integration-test-comp" 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        if echo "$output" | grep -q "successful patched"; then
            success "✓ Patch comprehensive command succeeded"
            echo "  $(echo "$output" | grep "successful patched")"
            return 0
        else
            error "✗ Patch comprehensive command output doesn't indicate success"
            echo "$output"
            return 1
        fi
    else
        error "✗ Patch comprehensive command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

test_patch_platforms_command() {
    log "Testing patch-platforms command..."
    local output
    set +e  # Don't exit on error for this test
    output=$($CLIENT_BINARY patch-platforms \
        --image "$TEST_IMAGE" \
        --patchtag "integration-test-plat" \
        --platform "linux/amd64" 2>&1)
    local exit_code=$?
    set -e  # Re-enable exit on error
    
    if [[ $exit_code -eq 0 ]]; then
        if echo "$output" | grep -q "successful patched"; then
            success "✓ Patch platforms command succeeded"
            echo "  $(echo "$output" | grep "successful patched")"
            return 0
        else
            error "✗ Patch platforms command output doesn't indicate success"
            echo "$output"
            return 1
        fi
    else
        error "✗ Patch platforms command failed with exit code $exit_code"
        echo "$output"
        return 1
    fi
}

# Test complete workflow
test_full_workflow() {
    log "Testing complete vulnerability-based patching workflow..."
    
    # Step 1: Scan
    log "  Step 1: Scanning image for vulnerabilities"
    if ! test_scan_command; then
        error "Workflow failed at scan step"
        return 1
    fi
    
    # Step 2: Patch based on scan
    log "  Step 2: Patching based on scan results"
    if ! test_patch_vulnerabilities_command; then
        error "Workflow failed at patch step"
        return 1
    fi
    
    success "✓ Complete workflow succeeded"
    return 0
}

# Test error scenarios
test_error_scenarios() {
    log "Testing error scenarios..."
    
    # Test with invalid image
    if $CLIENT_BINARY scan-container --image "invalid-image-12345" &> /dev/null; then
        warn "  Expected scan to fail with invalid image, but it succeeded"
    else
        log "  ✓ Scan correctly failed with invalid image"
    fi
    
    # Test with missing parameters
    if $CLIENT_BINARY patch-vulnerabilities --image "$TEST_IMAGE" &> /dev/null; then
        warn "  Expected patch-vulnerabilities to fail with missing parameters, but it succeeded"
    else
        log "  ✓ Patch-vulnerabilities correctly failed with missing parameters"
    fi
    
    success "✓ Error scenario tests completed"
}

# Cleanup function
cleanup() {
    if [[ -n "$REPORT_DIR" && -d "$REPORT_DIR" ]]; then
        log "Cleaning up report directory: $REPORT_DIR"
        rm -rf "$REPORT_DIR"
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Main test execution
main() {
    echo "=========================================="
    echo "  Copacetic MCP CLI Integration Tests"
    echo "=========================================="
    log "Test image: $TEST_IMAGE"
    log "Client binary: $CLIENT_BINARY"
    echo ""
    
    # Check prerequisites
    check_prerequisites
    echo ""
    
    test_count=0
    passed_count=0
    
    # Run individual command tests
    tests=(
        "test_version_command"
        "test_list_command"
        "test_scan_command"
        "test_patch_vulnerabilities_command"
        "test_patch_comprehensive_command"
        "test_patch_platforms_command"
    )
    
    for test in "${tests[@]}"; do
        test_count=$((test_count + 1))
        echo "----------------------------------------"
        log "Running test: $test"
        if $test; then
            passed_count=$((passed_count + 1))
        else
            error "Test failed: $test"
        fi
        echo ""
    done
    
    # Run workflow test
    test_count=$((test_count + 1))
    echo "----------------------------------------"
    log "Running complete workflow test"
    if test_full_workflow; then
        passed_count=$((passed_count + 1))
    else
        error "Workflow test failed"
    fi
    echo ""
    
    # Run error scenario tests
    echo "----------------------------------------"
    test_error_scenarios
    echo ""
    
    # Summary
    echo "=========================================="
    echo "           Test Summary"
    echo "=========================================="
    log "Passed: $passed_count/$test_count tests"
    
    if [[ $passed_count -eq $test_count ]]; then
        success "🎉 All integration tests passed!"
        echo ""
        exit 0
    else
        error "❌ Some tests failed."
        echo ""
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    "--help"|"-h")
        echo "Copacetic MCP CLI Integration Test"
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo "  --quick, -q   Run only quick tests (version, list)"
        echo ""
        echo "Environment variables:"
        echo "  TEST_IMAGE    Container image to test (default: alpine:3.17)"
        exit 0
        ;;
    "--quick"|"-q")
        echo "=========================================="
        echo "    Quick Integration Tests"
        echo "=========================================="
        check_prerequisites
        echo ""
        test_version_command
        echo ""
        test_list_command
        echo ""
        success "Quick tests completed successfully!"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac