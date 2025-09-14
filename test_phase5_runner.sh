#!/usr/bin/env bash

# Phase 5 Test Runner - Comprehensive Testing & Validation Framework
# Tests all somatogram percentile functionality, integration, and regression

set -e

echo "🧪 Phase 5: Running Comprehensive Testing & Validation Framework"
echo "============================================================="

cd "$(dirname "$0")"
PROJECT_ROOT="C:/Users/tranm/work/svw.info/portal64api"
KADER_PLANUNG_DIR="$PROJECT_ROOT/kader-planung"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Function to run test with timing and result tracking
run_test() {
    local test_name="$1"
    local test_command="$2"
    local test_description="$3"
    
    echo -e "${BLUE}Running: ${test_description}${NC}"
    echo "Command: $test_command"
    echo "----------------------------------------"
    
    local start_time=$(date +%s%N)
    
    if eval "$test_command"; then
        local end_time=$(date +%s%N)
        local duration=$(((end_time - start_time) / 1000000)) # Convert to milliseconds
        echo -e "${GREEN}✅ PASSED${NC} ($duration ms): $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAILED${NC}: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
}

# Change to kader-planung directory
cd "$KADER_PLANUNG_DIR"

echo "📍 Working Directory: $(pwd)"
echo "🚀 Starting Phase 5 Test Suite..."
echo ""

# 1. Unit Tests - Somatogram Percentile Logic
echo -e "${YELLOW}Phase 5.1: Unit Tests - Percentile Calculation Logic${NC}"
echo "====================================================="

run_test "somatogram_unit_tests" \
    "go test -v ./internal/processor -run TestGroupPlayersByAgeAndGender" \
    "Age-Gender Player Grouping Logic"

run_test "percentile_calculation" \
    "go test -v ./internal/processor -run TestCalculatePercentilesForGroup" \
    "Percentile Calculation Accuracy"

run_test "percentile_lookup" \
    "go test -v ./internal/processor -run TestFindPercentileForPlayer" \
    "Individual Player Percentile Lookup"

run_test "sample_size_filtering" \
    "go test -v ./internal/processor -run TestFilterGroupsBySampleSize" \
    "Minimum Sample Size Filtering"

run_test "percentile_integration" \
    "go test -v ./internal/processor -run TestCalculateSomatogramPercentilesIntegration" \
    "Complete Percentile Calculation Integration"

# 2. Integration Tests - Complete Pipeline
echo -e "${YELLOW}Phase 5.2: Integration Tests - Complete Pipeline${NC}"
echo "================================================="

run_test "pipeline_integration" \
    "go test -v ./internal/processor -run TestSomatogramIntegrationPipeline -timeout 60s" \
    "Complete Somatogram Pipeline Integration"

run_test "accuracy_validation" \
    "go test -v ./internal/processor -run TestSomatogramAccuracyValidation" \
    "Statistical Accuracy Validation"

run_test "performance_large_dataset" \
    "go test -v ./internal/processor -run TestSomatogramPerformanceWithLargeDataset -timeout 30s" \
    "Performance with Large Dataset (10k players)"

# 3. Export Tests - CSV Format with Somatogram Column
echo -e "${YELLOW}Phase 5.3: Export Tests - CSV with Somatogram Column${NC}"
echo "===================================================="

run_test "csv_export_somatogram" \
    "go test -v ./internal/export -run TestSomatogramPercentileCSVExport" \
    "CSV Export with Somatogram Percentile Column"

run_test "percentile_validation" \
    "go test -v ./internal/export -run TestSomatogramPercentileValidation" \
    "Percentile Value Format Validation"

run_test "export_edge_cases" \
    "go test -v ./internal/export -run TestSomatogramExportEdgeCases" \
    "Edge Cases in Somatogram Export"

run_test "german_csv_format" \
    "go test -v ./internal/export -run TestSomatogramCSVSemicolonSeparator" \
    "German Excel Compatibility (Semicolon Separator)"

# 4. Regression Tests - Backward Compatibility
echo -e "${YELLOW}Phase 5.4: Regression Tests - Backward Compatibility${NC}"
echo "===================================================="

run_test "csv_backward_compatibility" \
    "go test -v ./internal/processor -run TestBackwardCompatibilityCSVFormat" \
    "CSV Format Backward Compatibility"

run_test "existing_functionality" \
    "go test -v ./internal/processor -run TestExistingFunctionalityUnchanged" \
    "Existing Functionality Unchanged"

run_test "data_not_available_handling" \
    "go test -v ./internal/processor -run TestDataNotAvailableHandling" \
    "DATA_NOT_AVAILABLE Handling Consistency"

run_test "club_prefix_compatibility" \
    "go test -v ./internal/processor -run TestClubIDPrefixBackwardCompatibility" \
    "Club ID Prefix Functionality"

run_test "performance_baseline" \
    "go test -v ./internal/processor -run TestPerformanceRegressionBaseline -timeout 15s" \
    "Performance Regression Baseline"

run_test "parameter_cleanup_validation" \
    "go test -v ./internal/processor -run TestLegacyParameterCleanupValidation" \
    "Legacy Parameter Cleanup Validation"

# 5. Existing Test Suite Validation
echo -e "${YELLOW}Phase 5.5: Existing Test Suite Validation${NC}"
echo "=========================================="

run_test "existing_processor_tests" \
    "go test -v ./internal/processor -run TestAnalyzeHistoricalData" \
    "Existing Historical Analysis Tests"

run_test "existing_export_tests" \
    "go test -v ./internal/export -run TestExportCSV" \
    "Existing Export Functionality Tests"

run_test "models_tests" \
    "go test -v ./internal/models -run TestCalculateClubIDPrefixes" \
    "Models and Utility Function Tests"

# 6. Build and Compilation Tests
echo -e "${YELLOW}Phase 5.6: Build and Compilation Tests${NC}"
echo "======================================"

run_test "kader_planung_build" \
    "go build -o bin/test-kader-planung ./cmd/kader-planung/" \
    "Kader-Planung Application Build"

# Change to main Portal64 API directory for integration build test
cd "$PROJECT_ROOT"

run_test "portal64_api_build" \
    "./build.bat build" \
    "Portal64 API Integration Build"

# Test that built executable exists
if [ -f "bin/portal64api.exe" ]; then
    echo -e "${GREEN}✅ Portal64 API executable built successfully${NC}"
else
    echo -e "${RED}❌ Portal64 API executable not found${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# 7. Quick Smoke Test
echo -e "${YELLOW}Phase 5.7: Smoke Test - Application Startup${NC}"
echo "=============================================="

cd "$KADER_PLANUNG_DIR"

# Test that kader-planung binary works
if [ -f "bin/test-kader-planung" ]; then
    run_test "kader_planung_help" \
        "timeout 10s ./bin/test-kader-planung --help" \
        "Kader-Planung Help Command"
else
    echo -e "${RED}❌ Kader-Planung binary not found, skipping smoke test${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Cleanup test binary
rm -f "bin/test-kader-planung"

# 8. Test Summary and Results
echo -e "${BLUE}Phase 5 Test Results Summary${NC}"
echo "============================"
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED! Phase 5 Testing & Validation Framework Complete${NC}"
    echo ""
    echo "✅ Somatogram percentile calculation logic validated"
    echo "✅ Integration pipeline tested with realistic data"
    echo "✅ CSV export with new column verified"
    echo "✅ Backward compatibility confirmed"
    echo "✅ Performance benchmarks within acceptable limits"
    echo "✅ Edge cases and error conditions handled"
    echo ""
    echo "🚀 Phase 5: Testing & Validation Framework - COMPLETE"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please review and fix issues.${NC}"
    echo ""
    echo "Phase 5 Status: NEEDS ATTENTION"
    exit 1
fi
