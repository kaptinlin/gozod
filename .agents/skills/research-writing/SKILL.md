---
description: Generates structured research reports by analyzing reference projects in the .references directory. Use when the user requests research reports, feature analysis, or needs to analyze reference projects.
name: research-writing
---


# Research Writing

Generate structured research reports by analyzing reference projects in the `.references/` directory. Reports provide evidence-based analysis with actionable recommendations, API extractions, and interface design suggestions.

## When to Use

Use this skill when:
- User requests "生成调研报告" or "调研某特性"
- User asks to "generate research report" or "research a feature"
- Need to analyze reference implementations before writing specs
- Consolidating findings from multiple reference projects
- Building evidence base for design decisions

## Input Requirements

1. **References directory**: `.references/` containing reference projects (typically as git submodules)
2. **Research topics**: List of features/capabilities to investigate
3. **Output location**: `.research/` directory for generated reports

## Output Format

Reports are saved as `.research/R{XX}-{topic-slug}.md` (e.g., `.research/R01-wc-framework-core-apis.md`).

See `references/report-template.md` for the full template. Key structural elements:

```markdown
---
id: R__x
title: <报告标题>
task: <task ID>
date: YYYY-MM-DD
status: draft
scope: [...]
tags: [...]
---

# R__x — <报告标题>

## 执行摘要                    ← 结论表格 + 置信度
## 1-N. <主题章节>             ← 按主题横切多个项目，含 API 提炼
## N-3. 对本项目的落地建议      ← 文字 + 建议接口设计
## N-2. 决策矩阵               ← 推荐/备选/否决 + 依据
## N-1. 代码块索引              ← 快速定位所有 API/接口
## N. 引用清单                 ← 按项目分组
```

## Research Principles

- **Topic-driven**: Organize by topic dimension (横切多个项目), not by project card (每个项目一节)
- **Evidence-based**: Every recommendation must cite specific reference evidence. Use decision matrix format (推荐/备选/否决 + 依据)
- **API extraction + Suggested API**: Reports must include distilled API signatures from references AND proposed interfaces for our project. Types/signatures only, no implementation bodies
- **Language flexibility**: Write reports in the language matching the user's request. Keep technical terms, project names, and API names in original language

For heading hierarchy, source citation format, and code block rules, see `references/report-template.md`.

## Workflow

### Step 1: Identify Reference Projects
```bash
# List available references
ls -la .references/

# For submodules, check status
git submodule status
```

### Step 2: Analyze Reference Projects
For each reference project, examine:
- README.md (project overview, quick start)
- Documentation (design docs, architecture guides)
- **Source code** — focus on:
  - Public API surface (exported types, interfaces, function signatures)
  - Architecture patterns (how modules are organized)
  - Key abstractions (base classes, mixins, controllers)
- Configuration examples (YAML, JSON, code samples)

### Step 3: Extract Key Information
For each project, document:
- What problem it solves
- How it's designed (architecture approach)
- What features it provides (categorized)
- How to use it (configuration/code examples)
- What it can and cannot do (boundaries)

### Step 4: Synthesize by Topic
Organize findings by topic dimension, not by project:
- Identify common themes across projects
- Compare approaches to the same problem side-by-side
- Note different solutions, trade-offs, and patterns

### Step 5: Extract API Signatures
For key analysis points, distill API signatures:
- Extract core type/interface definitions from reference source
- Simplify to minimal signature (remove implementation noise)
- Add explanatory comments
- Place in the relevant topic section as "API 提炼" blocks

### Step 6: Design Suggested APIs
In the "落地建议" section:
- Convert text recommendations into interface drafts
- Include TypeScript types, CSS variable contracts, Schema structures
- Keep to signatures only — no function bodies

### Step 7: Build Decision Matrix
Create a decision matrix summarizing all key decisions:
- Recommended option with justification
- Alternative options with conditions
- Rejected options with reasons
- Reference to the section containing the analysis

### Step 8: Compile Reference List
Group all cited references by project:
- Include file path, line numbers, and usage purpose
- Ensure every inline `<!-- ref: ... -->` appears in the final list

### Step 9: Write Report File
**CRITICAL**: Write the complete report to file:
- Use path from user request if specified (e.g., ".research/R01-topic.md")
- Otherwise use standard path: `.research/R{XX}-{topic-slug}.md`
- Use Write tool to create the file
- Include all sections from the template
- Verify the file was created successfully

## File Organization

```
.research/
├── R01-{topic}.md            # Single-topic report (e.g., R01-wc-framework-core-apis.md)
├── R02-{topic}.md            # Another report
└── reports/                  # Completed phase reports (if applicable)
```

## Common Patterns

### Multi-Project Topic Analysis
When analyzing 3+ projects on the same topic:
1. Create one section per topic (not per project)
2. Compare approaches side-by-side within each section
3. Extract API signatures from the most representative projects
4. Build decision matrix at the end

### Depth Control
Match analysis depth to task requirements:
- **Feature survey**: Focus on capabilities and boundaries (docs + README)
- **Architecture research**: Deep-dive into source code for patterns and API extraction (default for most tasks)

## Integration with Other Skills

### Before spec-writing
Research reports provide evidence base for spec decisions. Reference the research report in spec rationale sections.

### After research-analyzing
The research-analyzing skill defines scope and generates ANALYSIS.md. This skill executes the research tasks and generates structured reports.

## Critical Requirements

**YOU MUST WRITE THE FILE**:
- If user specifies output path (e.g., "write .research/R01-topic.md"), use that exact path
- If no path specified, use standard path: `.research/R{XX}-{topic-slug}.md` (e.g., `.research/R01-token-system.md`)
- After completing analysis, use Write tool to create the file
- Do not just describe what should be written — actually write it
- Verify the file exists after writing

## Tips

- Organize by topic, not by project
- Read source code for API surfaces and patterns
- Focus on "what interface" and "why this design", not implementation details
- Include concrete API extractions (type signatures, not raw code)
- Every recommendation needs a suggested interface draft
- Use inline `<!-- ref: ... -->` for citations, keep prose clean
- Build decision matrix for every report with actionable decisions
