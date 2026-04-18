#!/usr/bin/env bash
# Tag a release with root + sub-module tags (same version).
# Pins ALL cross-module dependencies to the release version before tagging:
#   - sub-module → root
#   - root → sub-module
#   - sub-module → sub-module
#
# Usage: bash "$SKILL_SCRIPTS/tag_release.sh" <VERSION>  (where SKILL_SCRIPTS is resolved per SKILL.md)
# Example: bash "$SKILL_SCRIPTS/tag_release.sh" 1.3.0

set -euo pipefail

VERSION="${1:?Usage: tag_release.sh <VERSION> (e.g. 1.3.0)}"

# Normalize: strip leading 'v' if provided
VERSION="${VERSION#v}"

# Detect current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Guard: refuse to tag if tag already exists
if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
  echo "ERROR: Tag v${VERSION} already exists."
  exit 1
fi

# Discover sub-modules (exclude vendor, dotdirs, testdata)
SUBMODULES=$(find . -mindepth 2 -name go.mod \
    -not -path '*/vendor/*' \
    -not -path '*/.*' \
    -not -path '*/testdata/*' \
    -exec dirname {} \; | sed 's|^\./||' | sort)

# Collect all module paths in the repo
ROOT_MODULE=$(head -1 go.mod | awk '{print $2}')
declare -a ALL_MODULES=("$ROOT_MODULE")
if [ -n "$SUBMODULES" ]; then
  while IFS= read -r mod; do
    [ -z "$mod" ] && continue
    MOD_PATH=$(head -1 "${mod}/go.mod" | awk '{print $2}')
    ALL_MODULES+=("$MOD_PATH")
  done <<< "$SUBMODULES"
fi

# Collect all go.mod directories (root + sub-modules)
declare -a ALL_DIRS=(".")
if [ -n "$SUBMODULES" ]; then
  while IFS= read -r mod; do
    [ -z "$mod" ] && continue
    ALL_DIRS+=("$mod")
  done <<< "$SUBMODULES"
fi

# Pin ALL cross-module dependencies to release version
echo "Pinning cross-module dependencies to v${VERSION}..."
for dir in "${ALL_DIRS[@]}"; do
  SELF_MODULE=$(head -1 "${dir}/go.mod" | awk '{print $2}')
  for mod_path in "${ALL_MODULES[@]}"; do
    # Skip self-references
    [ "$mod_path" = "$SELF_MODULE" ] && continue
    # Pin if this go.mod references the module (matches both block and single-line require)
    if grep -q "${mod_path} v" "${dir}/go.mod" 2>/dev/null; then
      echo "  ${dir}/go.mod: ${mod_path} -> v${VERSION}"
      (cd "${dir}" && go mod edit -require="${mod_path}@v${VERSION}")
    fi
  done
done

# Stage all modified go.mod/go.sum files
git add go.mod go.sum 2>/dev/null || true
if [ -n "$SUBMODULES" ]; then
  while IFS= read -r mod; do
    [ -z "$mod" ] && continue
    git add "${mod}/go.mod" "${mod}/go.sum" 2>/dev/null || true
  done <<< "$SUBMODULES"
fi

# Only commit if there are staged changes
if ! git diff --cached --quiet; then
  git commit -m "build(deps): pin cross-module dependencies to v${VERSION}"
else
  echo "  (no changes to commit)"
fi

# Root tag
echo ""
echo "Creating tags..."
git tag -a "v${VERSION}" -m "v${VERSION}"
echo "  v${VERSION}"

# Sub-module tags (same version as root)
if [ -n "$SUBMODULES" ]; then
  while IFS= read -r mod; do
    [ -z "$mod" ] && continue
    git tag -a "${mod}/v${VERSION}" -m "${mod}/v${VERSION}"
    echo "  ${mod}/v${VERSION}"
  done <<< "$SUBMODULES"
fi

# Push current branch + tags
echo ""
echo "Pushing ${BRANCH} with tags..."
git push origin "${BRANCH}" --tags

echo ""
echo "Release v${VERSION} complete."
