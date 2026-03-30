#!/bin/bash
# Track metrics over time and save to CSV

DATE=$(date +%Y-%m-%d)
COMPLEXITY_COUNT=$(gocyclo . 2>/dev/null | wc -l | tr -d ' ')
DEAD_CODE_COUNT=$(deadcode ./... 2>/dev/null | wc -l | tr -d ' ')
COVERAGE=$(go test -cover ./... 2>&1 | grep -o '[0-9.]*%' | head -1 | tr -d '%')

# Append to history file
echo "$DATE,$COMPLEXITY_COUNT,$DEAD_CODE_COUNT,$COVERAGE" >> metrics-history.csv

echo "Metrics tracked: $DATE"
echo "  Complexity functions: $COMPLEXITY_COUNT"
echo "  Dead code items: $DEAD_CODE_COUNT"
echo "  Coverage: $COVERAGE%"
