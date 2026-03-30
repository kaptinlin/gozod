#!/usr/bin/env python3
"""
Validates skill metadata (name and description) against agentable skill standards.

Usage:
    python3 scripts/validate-metadata.py --name "skill-name" --description "Description text"
    python3 scripts/validate-metadata.py --file path/to/SKILL.md

Exit codes:
    0: Validation successful
    1: Validation failed (errors printed to stderr)
"""

import re
import sys
import os
import argparse
import yaml

def validate_metadata(name, description):
    """Validate skill name and description against standards."""
    errors = []
    warnings = []

    # 1. Validate Name Length
    if not (1 <= len(name) <= 64):
        errors.append(f"NAME ERROR: '{name}' is {len(name)} characters. Must be between 1-64.")

    # 2. Validate Name Characters (lowercase, numbers, single hyphens)
    # Regex: Starts/ends with alphanumeric, allows single hyphens in between
    if not re.match(r"^[a-z0-9]+(-[a-z0-9]+)*$", name):
        errors.append(
            f"NAME ERROR: '{name}' contains invalid characters. "
            "Use only lowercase letters, numbers, and single hyphens. "
            "No consecutive hyphens, and cannot start/end with a hyphen."
        )

    # 3. Validate Description Length
    if len(description) > 1024:
        errors.append(
            f"DESCRIPTION ERROR: Description is {len(description)} characters. "
            "Must be 1,024 characters or fewer."
        )

    # Warn if description is very short
    if len(description) < 50:
        warnings.append(
            f"DESCRIPTION WARNING: Description is only {len(description)} characters. "
            "Consider adding more context about when to use this skill."
        )

    # 4. Check for Third-Person Perspective (Basic Heuristic)
    first_person_words = {"i", "me", "my", "we", "our", "you", "your"}
    desc_words = set(re.findall(r'\b\w+\b', description.lower()))
    found_forbidden = first_person_words.intersection(desc_words)
    if found_forbidden:
        errors.append(
            f"STYLE ERROR: Description contains first/second person terms: {found_forbidden}. "
            "Use third-person imperative (e.g., 'Use when...', 'Creates...', 'Updates...')."
        )

    # 5. Check if description starts with "Use when" (recommended pattern)
    if not description.strip().startswith("Use when"):
        warnings.append(
            "STYLE WARNING: Description should start with 'Use when...' to focus on triggering conditions. "
            "See CSO (Claude Search Optimization) guidelines."
        )

    # 6. Check for workflow summary in description (anti-pattern)
    workflow_indicators = ["then", "after", "before", "first", "next", "finally", "step"]
    desc_lower = description.lower()
    found_workflow = [word for word in workflow_indicators if word in desc_lower]
    if len(found_workflow) >= 2:
        warnings.append(
            f"STYLE WARNING: Description may contain workflow summary (found: {found_workflow}). "
            "Description should describe WHEN to use, not HOW it works. "
            "See CSO guidelines about description vs. workflow."
        )

    # Print results
    if warnings:
        print("\n".join(warnings), file=sys.stderr)

    if errors:
        print("\n".join(errors), file=sys.stderr)
        sys.exit(1)
    else:
        print("✓ SUCCESS: Metadata is valid and optimized for discovery.")
        if warnings:
            print("  Note: Warnings above are recommendations, not errors.")
        sys.exit(0)

def extract_frontmatter(file_path):
    """Extract YAML frontmatter from SKILL.md file."""
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Match YAML frontmatter between --- delimiters
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
    if not match:
        print("ERROR: No YAML frontmatter found in file.", file=sys.stderr)
        sys.exit(1)

    frontmatter = yaml.safe_load(match.group(1))

    if 'name' not in frontmatter or 'description' not in frontmatter:
        print("ERROR: Frontmatter must contain 'name' and 'description' fields.", file=sys.stderr)
        sys.exit(1)

    return frontmatter['name'], frontmatter['description']

def validate_file_references(skill_md_path):
    """Check if referenced files exist and return missing references."""
    skill_dir = os.path.dirname(skill_md_path)
    if not skill_dir:
        skill_dir = '.'

    with open(skill_md_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Find all references like `references/file.md`, `scripts/file.py`, `assets/file.md`
    refs = re.findall(r'`((?:references|scripts|assets)/[^`]+)`', content)

    # Filter out obvious placeholders
    placeholders = ['[file]', '[script-name]', '[doc-name]', '[template-name]',
                   'api-spec.md', 'troubleshooting.md', 'testing-methodology.md']

    missing = []
    for ref in set(refs):  # Use set to avoid duplicates
        # Skip if it's a placeholder
        if any(placeholder in ref for placeholder in placeholders):
            continue

        ref_path = os.path.join(skill_dir, ref)
        if not os.path.exists(ref_path):
            missing.append(ref)

    return missing

def check_line_count(skill_md_path):
    """Check SKILL.md line count and return warning if over 500 lines."""
    with open(skill_md_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    line_count = len(lines)
    if line_count > 500:
        return f"LINE COUNT WARNING: SKILL.md has {line_count} lines. Target is <500 lines. Consider extracting content to references/."

    return None

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Validate skill metadata against agentable standards."
    )
    parser.add_argument(
        "--name",
        help="Skill name to validate"
    )
    parser.add_argument(
        "--description",
        help="Skill description to validate"
    )
    parser.add_argument(
        "--file",
        help="Path to SKILL.md file (extracts frontmatter automatically)"
    )

    args = parser.parse_args()

    # Validate input
    if args.file:
        name, description = extract_frontmatter(args.file)

        # Additional file-based validations
        missing_refs = validate_file_references(args.file)
        if missing_refs:
            print(f"\nFILE REFERENCE WARNING: The following referenced files do not exist:", file=sys.stderr)
            for ref in missing_refs:
                print(f"  - {ref}", file=sys.stderr)

        line_warning = check_line_count(args.file)
        if line_warning:
            print(f"\n{line_warning}", file=sys.stderr)

    elif args.name and args.description:
        name = args.name
        description = args.description
    else:
        parser.print_help()
        print("\nERROR: Must provide either --file or both --name and --description", file=sys.stderr)
        sys.exit(1)

    validate_metadata(name, description)
