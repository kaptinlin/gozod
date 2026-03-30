#!/usr/bin/env python3
"""
Skill Initializer - Creates a new skill from template

Usage:
    init_skill.py <skill-name> [--type TYPE] [--path PATH]

Examples:
    init_skill.py my-new-skill
    init_skill.py my-technique --type technique
    init_skill.py my-pattern --type pattern --path /custom/location
"""

import sys
import subprocess
from pathlib import Path


def title_case_skill_name(skill_name):
    """Convert hyphenated skill name to Title Case for display."""
    return ' '.join(word.capitalize() for word in skill_name.split('-'))


def validate_skill_name(skill_name):
    """Validate skill name using validate-metadata.py."""
    try:
        # Find the validate-metadata.py script relative to this script
        script_dir = Path(__file__).parent
        validate_script = script_dir / 'validate-metadata.py'

        if not validate_script.exists():
            print(f"⚠️  Warning: Validation script not found at {validate_script}")
            return True  # Continue anyway

        # Run validation with a placeholder description
        result = subprocess.run(
            ['python3', str(validate_script),
             '--name', skill_name,
             '--description', 'Placeholder description for validation'],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            print("❌ Skill name validation failed:")
            print(result.stderr)
            return False

        return True
    except Exception as e:
        print(f"⚠️  Warning: Could not run validation: {e}")
        return True  # Continue anyway


def get_template_path(skill_type):
    """Get the appropriate template file based on skill type."""
    # Find templates relative to this script
    script_dir = Path(__file__).parent
    skill_creating_dir = script_dir.parent

    templates = {
        'general': skill_creating_dir / 'assets/skill-template.md',
        'technique': skill_creating_dir / 'assets/skill-template-technique.md',
        'pattern': skill_creating_dir / 'assets/skill-template-pattern.md',
        'reference': skill_creating_dir / 'assets/skill-template-reference.md'
    }

    template = templates.get(skill_type, templates['general'])

    if not template.exists():
        print(f"⚠️  Warning: Template not found: {template}")
        print(f"   Using general template instead.")
        return templates['general']

    return template


def init_skill(skill_name, skill_type='general', path='.'):
    """
    Initialize a new skill directory with template SKILL.md.

    Args:
        skill_name: Name of the skill
        skill_type: Type of skill (general, technique, pattern, reference)
        path: Path where the skill directory should be created

    Returns:
        Path to created skill directory, or None if error
    """
    # Validate skill name first
    if not validate_skill_name(skill_name):
        return None

    # Determine skill directory path
    skill_dir = Path(path).resolve() / skill_name

    # Check if directory already exists
    if skill_dir.exists():
        print(f"❌ Error: Skill directory already exists: {skill_dir}")
        return None

    # Create skill directory
    try:
        skill_dir.mkdir(parents=True, exist_ok=False)
        print(f"✅ Created skill directory: {skill_dir}")
    except Exception as e:
        print(f"❌ Error creating directory: {e}")
        return None

    # Get appropriate template
    template_path = get_template_path(skill_type)

    # Copy template to SKILL.md
    skill_md_path = skill_dir / 'SKILL.md'
    try:
        template_content = template_path.read_text()

        # Replace placeholders
        skill_title = title_case_skill_name(skill_name)
        template_content = template_content.replace('[skill-name]', skill_name)
        template_content = template_content.replace('[Skill Title]', skill_title)
        template_content = template_content.replace('[Technique Name]', skill_title)
        template_content = template_content.replace('[Pattern Name]', skill_title)
        template_content = template_content.replace('[API/Tool/Library Name]', skill_title)

        skill_md_path.write_text(template_content)
        print(f"✅ Created SKILL.md from {skill_type} template")
    except Exception as e:
        print(f"❌ Error creating SKILL.md: {e}")
        return None

    # Create resource directories
    try:
        # Create scripts/ directory
        scripts_dir = skill_dir / 'scripts'
        scripts_dir.mkdir(exist_ok=True)
        print("✅ Created scripts/ directory")

        # Create references/ directory
        references_dir = skill_dir / 'references'
        references_dir.mkdir(exist_ok=True)
        print("✅ Created references/ directory")

        # Create assets/ directory
        assets_dir = skill_dir / 'assets'
        assets_dir.mkdir(exist_ok=True)
        print("✅ Created assets/ directory")

    except Exception as e:
        print(f"❌ Error creating resource directories: {e}")
        return None

    # Print next steps
    print(f"\n✅ Skill '{skill_name}' initialized successfully at {skill_dir}")
    print(f"   Type: {skill_type}")
    print("\nNext steps:")
    print("1. Edit SKILL.md to complete the placeholders:")
    print("   - Update description with specific triggers and use cases")
    print("   - Fill in all [placeholder] sections")
    print("   - Add concrete examples and code")
    print("2. Validate metadata:")
    print(f"   python3 scripts/validate-metadata.py --file {skill_dir}/SKILL.md")
    print("3. Add supporting files to scripts/, references/, or assets/ as needed")
    print("4. Test the skill with subagents before deployment")
    print("5. Review the checklist:")
    print("   references/skill-creation-checklist.md")

    return skill_dir


def main():
    # Parse arguments
    if len(sys.argv) < 2:
        print("Usage: init_skill.py <skill-name> [--type TYPE] [--path PATH]")
        print("\nSkill name requirements:")
        print("  - Lowercase letters, numbers, and single hyphens only")
        print("  - 1-64 characters")
        print("  - No consecutive hyphens, cannot start/end with hyphen")
        print("\nSkill types:")
        print("  - general (default): Generic skill template")
        print("  - technique: Step-by-step how-to guide")
        print("  - pattern: Mental model or way of thinking")
        print("  - reference: API docs, syntax guides, tool documentation")
        print("\nExamples:")
        print("  init_skill.py my-new-skill")
        print("  init_skill.py async-testing --type technique")
        print("  init_skill.py flatten-with-flags --type pattern")
        print("  init_skill.py pptx-api --type reference --path /custom/location")
        sys.exit(1)

    skill_name = sys.argv[1]
    skill_type = 'general'
    path = '.'

    # Parse optional arguments
    i = 2
    while i < len(sys.argv):
        if sys.argv[i] == '--type' and i + 1 < len(sys.argv):
            skill_type = sys.argv[i + 1]
            i += 2
        elif sys.argv[i] == '--path' and i + 1 < len(sys.argv):
            path = sys.argv[i + 1]
            i += 2
        else:
            print(f"Unknown argument: {sys.argv[i]}")
            sys.exit(1)

    # Validate skill type
    valid_types = ['general', 'technique', 'pattern', 'reference']
    if skill_type not in valid_types:
        print(f"❌ Error: Invalid skill type '{skill_type}'")
        print(f"   Valid types: {', '.join(valid_types)}")
        sys.exit(1)

    print(f"🚀 Initializing skill: {skill_name}")
    print(f"   Type: {skill_type}")
    print(f"   Location: {path}")
    print()

    result = init_skill(skill_name, skill_type, path)

    if result:
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()
