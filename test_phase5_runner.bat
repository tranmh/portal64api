@echo off
REM Phase 5 Test Runner - Comprehensive Testing & Validation Framework
REM Tests all somatogram percentile functionality, integration, and regression

setlocal enabledelayedexpansion

echo 🧪 Phase 5: Running Comprehensive Testing ^& Validation Framework
echo =============================================================

set PROJECT_ROOT=C:\Users\tranm\work\svw.info\portal64api
set KADER_PLANUNG_DIR=%PROJECT_ROOT%\kader-planung

REM Test result tracking
set TESTS_PASSED=0
set TESTS_FAILED=0
set TOTAL_TESTS=0

REM Change to kader-planung directory
cd /d "%KADER_PLANUNG_DIR%"

echo 📍 Working Directory: %cd%
echo 🚀 Starting Phase 5 Test Suite...
echo.

REM Function to run test with result tracking (Windows batch version)
goto :start_tests

:run_test
set test_name=%~1
set test_command=%~2
set test_description=%~3

echo Running: %test_description%
echo Command: %test_command%
echo ----------------------------------------

REM Execute the command and capture return code
%test_command%
if !ERRORLEVEL! == 0 (
    echo ✅ PASSED: %test_name%
    set /a TESTS_PASSED+=1
) else (
    echo ❌ FAILED: %test_name%
    set /a TESTS_FAILED+=1
)

set /a TOTAL_TESTS+=1
echo.
goto :eof

:start_tests

REM 1. Unit Tests - Somatogram Percentile Logic
echo Phase 5.1: Unit Tests - Percentile Calculation Logic
echo =====================================================

call :run_test "somatogram_unit_tests" "go test -v ./internal/processor -run TestGroupPlayersByAgeAndGender" "Age-Gender Player Grouping Logic"

call :run_test "percentile_calculation" "go test -v ./internal/processor -run TestCalculatePercentilesForGroup" "Percentile Calculation Accuracy"

call :run_test "percentile_lookup" "go test -v ./internal/processor -run TestFindPercentileForPlayer" "Individual Player Percentile Lookup"

call :run_test "sample_size_filtering" "go test -v ./internal/processor -run TestFilterGroupsBySampleSize" "Minimum Sample Size Filtering"

call :run_test "percentile_integration" "go test -v ./internal/processor -run TestCalculateSomatogramPercentilesIntegration" "Complete Percentile Calculation Integration"

REM 2. Integration Tests - Complete Pipeline
echo Phase 5.2: Integration Tests - Complete Pipeline
echo =================================================

call :run_test "pipeline_integration" "go test -v ./internal/processor -run TestSomatogramIntegrationPipeline -timeout 60s" "Complete Somatogram Pipeline Integration"

call :run_test "accuracy_validation" "go test -v ./internal/processor -run TestSomatogramAccuracyValidation" "Statistical Accuracy Validation"

call :run_test "performance_large_dataset" "go test -v ./internal/processor -run TestSomatogramPerformanceWithLargeDataset -timeout 30s" "Performance with Large Dataset (10k players)"

REM 3. Export Tests - CSV Format with Somatogram Column
echo Phase 5.3: Export Tests - CSV with Somatogram Column
echo ====================================================

call :run_test "csv_export_somatogram" "go test -v ./internal/export -run TestSomatogramPercentileCSVExport" "CSV Export with Somatogram Percentile Column"

call :run_test "percentile_validation" "go test -v ./internal/export -run TestSomatogramPercentileValidation" "Percentile Value Format Validation"

call :run_test "export_edge_cases" "go test -v ./internal/export -run TestSomatogramExportEdgeCases" "Edge Cases in Somatogram Export"

call :run_test "german_csv_format" "go test -v ./internal/export -run TestSomatogramCSVSemicolonSeparator" "German Excel Compatibility (Semicolon Separator)"

REM 4. Regression Tests - Backward Compatibility
echo Phase 5.4: Regression Tests - Backward Compatibility
echo ====================================================

call :run_test "csv_backward_compatibility" "go test -v ./internal/processor -run TestBackwardCompatibilityCSVFormat" "CSV Format Backward Compatibility"

call :run_test "existing_functionality" "go test -v ./internal/processor -run TestExistingFunctionalityUnchanged" "Existing Functionality Unchanged"

call :run_test "data_not_available_handling" "go test -v ./internal/processor -run TestDataNotAvailableHandling" "DATA_NOT_AVAILABLE Handling Consistency"

call :run_test "club_prefix_compatibility" "go test -v ./internal/processor -run TestClubIDPrefixBackwardCompatibility" "Club ID Prefix Functionality"

call :run_test "performance_baseline" "go test -v ./internal/processor -run TestPerformanceRegressionBaseline -timeout 15s" "Performance Regression Baseline"

call :run_test "parameter_cleanup_validation" "go test -v ./internal/processor -run TestLegacyParameterCleanupValidation" "Legacy Parameter Cleanup Validation"

REM 5. Existing Test Suite Validation
echo Phase 5.5: Existing Test Suite Validation
echo ==========================================

call :run_test "existing_processor_tests" "go test -v ./internal/processor -run TestAnalyzeHistoricalData" "Existing Historical Analysis Tests"

call :run_test "existing_export_tests" "go test -v ./internal/export -run TestExportCSV" "Existing Export Functionality Tests"

call :run_test "models_tests" "go test -v ./internal/models -run TestCalculateClubIDPrefixes" "Models and Utility Function Tests"

REM 6. Build and Compilation Tests
echo Phase 5.6: Build and Compilation Tests
echo ======================================

call :run_test "kader_planung_build" "go build -o bin/test-kader-planung.exe ./cmd/kader-planung/" "Kader-Planung Application Build"

REM Change to main Portal64 API directory for integration build test
cd /d "%PROJECT_ROOT%"

call :run_test "portal64_api_build" ".\build.bat build" "Portal64 API Integration Build"

REM Test that built executable exists
if exist "bin\portal64api.exe" (
    echo ✅ Portal64 API executable built successfully
) else (
    echo ❌ Portal64 API executable not found
    set /a TESTS_FAILED+=1
)

REM 7. Quick Smoke Test
echo Phase 5.7: Smoke Test - Application Startup
echo ==============================================

cd /d "%KADER_PLANUNG_DIR%"

REM Test that kader-planung binary works
if exist "bin\test-kader-planung.exe" (
    call :run_test "kader_planung_help" "bin\test-kader-planung.exe --help" "Kader-Planung Help Command"
) else (
    echo ❌ Kader-Planung binary not found, skipping smoke test
    set /a TESTS_FAILED+=1
)

REM Cleanup test binary
if exist "bin\test-kader-planung.exe" del "bin\test-kader-planung.exe"

REM 8. Test Summary and Results
echo Phase 5 Test Results Summary
echo ============================
echo Total Tests: %TOTAL_TESTS%
echo Passed: %TESTS_PASSED%
echo Failed: %TESTS_FAILED%

if %TESTS_FAILED% == 0 (
    echo.
    echo 🎉 ALL TESTS PASSED! Phase 5 Testing ^& Validation Framework Complete
    echo.
    echo ✅ Somatogram percentile calculation logic validated
    echo ✅ Integration pipeline tested with realistic data
    echo ✅ CSV export with new column verified
    echo ✅ Backward compatibility confirmed
    echo ✅ Performance benchmarks within acceptable limits
    echo ✅ Edge cases and error conditions handled
    echo.
    echo 🚀 Phase 5: Testing ^& Validation Framework - COMPLETE
    exit /b 0
) else (
    echo.
    echo ❌ Some tests failed. Please review and fix issues.
    echo.
    echo Phase 5 Status: NEEDS ATTENTION
    exit /b 1
)
