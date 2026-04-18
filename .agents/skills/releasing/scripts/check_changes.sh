#!/usr/bin/env bash
# Check for changes since the latest tag that warrant a new release.
# Exit 0 = tag needed, Exit 1 = no tag needed.
#
# Usage: bash "$SKILL_SCRIPTS/check_changes.sh"  (where SKILL_SCRIPTS is resolved per SKILL.md)

set -euo pipefail

LATEST_TAG=$(git tag --list 'v*' --sort=-v:refname | head -1)

if [ -z "$LATEST_TAG" ]; then
  echo "No existing tags found — first release."
  exit 0
fi

echo "Latest tag: $LATEST_TAG"
echo ""

TAG_NEEDED=false

# 1. Go source file changes
GO_CHANGED=$(git diff "$LATEST_TAG"..HEAD --name-only -- '*.go' || true)
if [ -n "$GO_CHANGED" ]; then
  echo "Go file changes:"
  echo "$GO_CHANGED" | sed 's/^/  /'
  echo ""
  TAG_NEEDED=true
fi

# 2. Root go.mod changes (any require/replace/retract lines changed)
#    Uses full diff to detect both grouped and single-line directives.
GOMOD_CHANGED=$(git diff "$LATEST_TAG"..HEAD -- go.mod | \
  grep -E '^[+-]' | \
  grep -v -E '^(\+\+\+|---) ' | \
  grep -v -E '^[+-](module |go |toolchain )' | \
  grep -v -E '^[+-]\)$' || true)
if [ -n "$GOMOD_CHANGED" ]; then
  echo "Root go.mod dependency changes:"
  echo "$GOMOD_CHANGED" | sed 's/^/  /'
  echo ""
  TAG_NEEDED=true
fi

# 3. Root go.sum changes
GOSUM_CHANGED=$(git diff "$LATEST_TAG"..HEAD --name-only -- go.sum || true)
if [ -n "$GOSUM_CHANGED" ]; then
  echo "Root go.sum changes detected."
  echo ""
  TAG_NEEDED=true
fi

# 4. Sub-module go.mod/go.sum changes
SUBMOD_CHANGED=$(git diff "$LATEST_TAG"..HEAD --name-only -- '*/go.mod' '*/go.sum' || true)
if [ -n "$SUBMOD_CHANGED" ]; then
  echo "Sub-module dependency changes:"
  echo "$SUBMOD_CHANGED" | sed 's/^/  /'
  echo ""
  TAG_NEEDED=true
fi

if [ "$TAG_NEEDED" = true ]; then
  echo "Result: TAG NEEDED"
  exit 0
else
  echo "No Go source or dependency changes since $LATEST_TAG."
  echo "Result: NO TAG NEEDED (push without tagging)"
  exit 1
fi
