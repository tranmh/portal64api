#!/bin/bash
#
# Kader-Planung Daily Execution Script
# This script runs kader-planung directly to generate CSV reports
#

# Configuration
BINARY="/opt/portal64api/kader-planung/bin/kader-planung"
OUTPUT_DIR="/opt/portal64api/internal/static/demo/kader-planung"
API_BASE_URL="https://test.svw.info:8080"
LOG_DIR="/opt/portal64api/logs"
LOG_FILE="$LOG_DIR/kader-planung-$(date +%Y%m%d-%H%M%S).log"

# Ensure log directory exists
mkdir -p "$LOG_DIR"

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"

# Run kader-planung
echo "========================================" >> "$LOG_FILE"
echo "Kader-Planung Execution Started: $(date)" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

cd "$OUTPUT_DIR" || exit 1

"$BINARY" \
  --api-base-url "$API_BASE_URL" \
  --output-dir "$OUTPUT_DIR" \
  --concurrency 16 \
  --min-sample-size 10 \
  --timeout 60 \
  --verbose \
  >> "$LOG_FILE" 2>&1

EXIT_CODE=$?

echo "" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"
echo "Kader-Planung Execution Finished: $(date)" >> "$LOG_FILE"
echo "Exit Code: $EXIT_CODE" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

# List generated files
echo "Generated files:" >> "$LOG_FILE"
ls -lh "$OUTPUT_DIR"/*.csv 2>/dev/null | tail -5 >> "$LOG_FILE"

# Keep only the last 10 CSV files (cleanup old ones)
cd "$OUTPUT_DIR" && ls -t kader-planung-*.csv 2>/dev/null | tail -n +11 | xargs -r rm -f

# Keep only the last 30 log files
cd "$LOG_DIR" && ls -t kader-planung-*.log 2>/dev/null | tail -n +31 | xargs -r rm -f

# Send summary to syslog
if [ $EXIT_CODE -eq 0 ]; then
    logger -t kader-planung "SUCCESS: Daily execution completed"
else
    logger -t kader-planung "ERROR: Daily execution failed with exit code $EXIT_CODE"
fi

exit $EXIT_CODE
