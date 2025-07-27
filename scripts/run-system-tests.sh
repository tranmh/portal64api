#!/bin/bash
# Bash script to run system tests for Portal64 API

set -e

# Default values
BASE_URL="${PORTAL64_TEST_BASE_URL:-http://test.svw.info:8080}"
VERBOSE=false
RACE=false
TIMEOUT=300

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --race)
            RACE=true
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --base-url URL    Base URL for the API (default: http://test.svw.info:8080)"
            echo "  --verbose, -v     Enable verbose output"
            echo "  --race            Enable race detection"
            echo "  --timeout SEC     Test timeout in seconds (default: 300)"
            echo "  --help, -h        Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${GREEN}Portal64 API System Tests${NC}"
echo -e "${GREEN}=========================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# Check server health
echo -e "${YELLOW}Checking server health at $BASE_URL...${NC}"
if curl -f -s "$BASE_URL/health" > /dev/null; then
    echo -e "${GREEN}✓ Server is healthy${NC}"
else
    echo -e "${RED}✗ Cannot reach server at $BASE_URL${NC}"
    echo ""
    echo -e "${YELLOW}Please ensure:${NC}"
    echo -e "${YELLOW}  1. The server is running at $BASE_URL${NC}"
    echo -e "${YELLOW}  2. Your network connection is working${NC}"
    echo -e "${YELLOW}  3. No firewall is blocking the connection${NC}"
    exit 1
fi

# Build test command
TEST_CMD="go test ./tests/integration/ -run TestSystemSuite -timeout=${TIMEOUT}s"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$RACE" = true ]; then
    TEST_CMD="$TEST_CMD -race"
fi

echo ""
echo -e "${GREEN}Running system tests...${NC}"
echo -e "${CYAN}Command: $TEST_CMD${NC}"
echo ""

# Export base URL for tests
export PORTAL64_TEST_BASE_URL="$BASE_URL"

# Run the tests
if eval $TEST_CMD; then
    echo ""
    echo -e "${GREEN}✓ All system tests passed!${NC}"
    exit 0
else
    EXIT_CODE=$?
    echo ""
    echo -e "${RED}✗ Some system tests failed (exit code: $EXIT_CODE)${NC}"
    echo ""
    echo -e "${YELLOW}Troubleshooting tips:${NC}"
    echo -e "${YELLOW}  1. Check server logs for errors${NC}"
    echo -e "${YELLOW}  2. Verify test data exists in the database${NC}"
    echo -e "${YELLOW}  3. Run with --verbose flag for detailed output${NC}"
    echo -e "${YELLOW}  4. Check network connectivity to the test server${NC}"
    exit $EXIT_CODE
fi
