#!/bin/bash
# Quick metrics collection for architecture audit

echo "=== Package Sizes (LOC) ==="
for pkg in pkg/*; do
  if [ -d "$pkg" ]; then
    loc=$(find "$pkg" -name "*.go" -not -name "*_test.go" -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
    echo "$(basename $pkg) $loc"
  fi
done | sort -k2 -rn | head -10

echo ""
echo "=== Package Coupling (Imports) ==="
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./pkg/... 2>/dev/null | \
  awk '{print $1, NF-1}' | sort -k2 -rn | head -10

echo ""
echo "=== Test Coverage ==="
go test -cover ./... 2>&1 | grep -E "coverage:" | head -10
