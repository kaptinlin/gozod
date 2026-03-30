---
name: write-file
description: Use when creating new files with the Write tool or encountering "Error writing file" — covers required parameters, absolute path usage, directory creation, multi-line content, chunking for large files, error troubleshooting, and common Agent mistakes. Don't use for editing existing files (use Edit tool), reading files, or non-file-writing tasks.
---

# Write File Skill

## Overview

Correctly use the Write tool to create new files. The Write tool requires two mandatory parameters and has specific rules for path handling, content formatting, and large file chunking.

## When to Use

**Use this skill when:**
- Creating new source files in any language
- Creating new configuration files (JSON, YAML, TOML, etc.)
- Creating new documentation files (Markdown, AsciiDoc, etc.)
- Creating new test files
- Unsure about Write tool parameter structure
- Getting "missing parameter" errors from Write tool
- Encountering "Error writing file" errors

**Don't use this skill for:**
- Editing existing files (use Edit tool instead)
- Reading files (use Read tool)
- Searching for files (use Glob/Grep)
- Running shell commands (use Bash tool)

## Core Rules

### Required Parameters

```typescript
Write({
  file_path: string,  // Required: absolute path
  content: string     // Required: file content
})
```

**Key Points:**
- ✅ Create new files with Write, modify existing files with Edit
- ✅ Always use absolute paths (`~/work/...` or `/absolute/path`)
- ✅ Ensure parent directory exists (`mkdir -p`)
- ✅ Use template strings (backticks) for multi-line content
- ❌ Don't use empty object `{}`
- ❌ Don't use `input` nested parameter
- ❌ Don't omit any required parameter

## Correct Examples

### Example 1: TypeScript File

```typescript
Write({
  file_path: "/path/to/project/src/utils.ts",
  content: `export function calculateTotal(items: number[]): number {
  return items.reduce((sum, item) => sum + item, 0);
}

export function formatCurrency(amount: number): string {
  return \`$\${amount.toFixed(2)}\`;
}
`
})
```

### Example 2: JSON Configuration

```typescript
Write({
  file_path: "/path/to/config.json",
  content: `{"name": "my-project", "version": "1.0.0", "type": "module"}`
})
```

### Example 3: Markdown Document

```typescript
Write({
  file_path: "/path/to/docs/guide.md",
  content: `# User Guide

## Getting Started
This guide covers the basics...
`
})
```

## Common Mistakes

### Mistake 1: Empty Object

```typescript
// ❌ Wrong: empty object
Write({})

// ✅ Correct: both parameters provided
Write({
  file_path: "~/work/project/file.md",
  content: "Actual content"
})
```

### Mistake 2: Missing Parameters

```typescript
// ❌ Wrong: missing content
Write({ file_path: "~/work/project/file.md" })

// ❌ Wrong: missing file_path
Write({ content: "Content" })

// ✅ Correct: both parameters
Write({
  file_path: "~/work/project/file.md",
  content: "Content"
})
```

### Mistake 3: Wrong Parameter Structure

```typescript
// ❌ Wrong: nested structure
Write({ input: { file_path: "~/work/...", content: "..." } })

// ✅ Correct: flat structure
Write({ file_path: "~/work/...", content: "..." })
```

### Mistake 4: Relative Path

```typescript
// ❌ Not recommended: relative path
Write({ file_path: "./relative/path.md", content: "..." })

// ✅ Recommended: absolute path
Write({ file_path: "~/work/project/path.md", content: "..." })
```

### Mistake 5: Parent Directory Doesn't Exist

```typescript
// ❌ Wrong: directory doesn't exist
Write({ file_path: "~/work/nonexistent-dir/file.md", content: "..." })

// ✅ Correct: create directory first
Bash({ command: "mkdir -p ~/work/new-dir", description: "Create directory" })
Write({ file_path: "~/work/new-dir/file.md", content: "..." })
```

## Best Practices

### 1. Ensure Directory Exists

```typescript
Bash({ command: "mkdir -p ~/work/project/dir", description: "Create directory" })
Write({ file_path: "~/work/project/dir/file.md", content: "..." })
```

### 2. Overwriting Existing Files Requires Read First

```typescript
Read({ file_path: "~/work/project/existing.md" })
Write({ file_path: "~/work/project/existing.md", content: "New content" })
```

### 3. Use Edit for Modifying Existing Files

```typescript
// ❌ Wrong: overwrites entire file
Write({ file_path: "~/work/project/file.md", content: "New content" })

// ✅ Correct: modifies specific content
Edit({ file_path: "~/work/project/file.md", old_string: "Old", new_string: "New" })
```

### 4. Chunk Large Files (>150 lines)

```typescript
// First chunk
Write({
  file_path: "~/work/project/large.md",
  content: `# Part 1
...
// __CONTINUE__`
})

// Second chunk
Edit({
  file_path: "~/work/project/large.md",
  old_string: "// __CONTINUE__",
  new_string: `# Part 2
...`
})
```

## Implementation Steps

### Step 1: Verify Parent Directory

```bash
# Check if directory exists
ls -la ~/work/project/dir

# Create if needed
mkdir -p ~/work/project/dir
```

### Step 2: Prepare Content

- Use template strings (backticks) for multi-line content
- Escape special characters if needed
- Keep content under 150 lines (chunk if larger)

### Step 3: Call Write Tool

```typescript
Write({
  file_path: "~/work/project/dir/file.md",
  content: `Multi-line
content
here`
})
```

### Step 4: Verify File Created

```bash
# Check file exists
ls -la ~/work/project/dir/file.md

# View content
cat ~/work/project/dir/file.md
```

## Debugging Checklist

| Error | Cause | Solution |
|-------|-------|----------|
| `missing file_path` | Missing file_path parameter | Add `file_path: "~/work/..."` |
| `missing content` | Missing content parameter | Add `content: "..."` |
| `ENOENT` | Parent directory doesn't exist | Run `mkdir -p ~/work/parent/dir` first |
| `EACCES` | Permission denied | Check path permissions, avoid system directories |
| `Error writing file` | Various causes (see below) | Check file path, directory permissions, content format |

### Error writing file

When you encounter "Error writing file", check these common causes:

1. **Parent directory doesn't exist**
   ```bash
   # Create parent directory first
   mkdir -p ~/work/project/subdir
   ```

2. **Invalid file path**
   ```typescript
   // ❌ Wrong: contains invalid characters or malformed path
   Write({ file_path: "~/work/project//double-slash.md", content: "..." })

   // ✅ Correct: clean path
   Write({ file_path: "~/work/project/file.md", content: "..." })
   ```

3. **Permission issues**
   ```bash
   # Check directory permissions
   ls -la ~/work/project

   # Fix permissions if needed
   chmod 755 ~/work/project
   ```

4. **File already exists and wasn't read first**
   ```typescript
   // ❌ Wrong: overwriting without reading first
   Write({ file_path: "~/work/existing.md", content: "..." })

   // ✅ Correct: read before overwriting
   Read({ file_path: "~/work/existing.md" })
   Write({ file_path: "~/work/existing.md", content: "..." })
   ```

5. **Content too large (>150 lines)**
   - Solution: Use chunking approach (see "Chunk Large Files" section)

## Common Agent Errors

### Error Pattern 1: Undefined Parameters

```typescript
// ❌ Agent common error: parameters exist but values are undefined
Write({
  file_path: undefined,
  content: undefined
})

// ✅ Correct: ensure parameters have actual values
Write({
  file_path: "~/work/project/file.md",
  content: "# Actual Content\n\nReal content here."
})
```

### Error Pattern 2: Using Wrong Tool

```typescript
// ❌ Wrong: using Write to modify existing file
Write({ file_path: "~/work/existing.md", content: "New content" })

// ✅ Correct: use Edit for modifications
Edit({ file_path: "~/work/existing.md", old_string: "Old", new_string: "New" })
```

## Supporting Files

- System prompt: Write tool usage instructions
- CLAUDE.md: Project-specific file creation conventions
- `.agents/skills/`: General-purpose skills for other file operations
