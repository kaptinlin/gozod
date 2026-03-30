# SPEC 结构

## Overview

定义 SPEC 的标准章节、顺序和内联决策记录格式。

不涵盖目录组织（见 [10-organization.md]）、命名规范（见 [12-naming.md]）、内容边界（见 [20-content-boundaries.md]）。

## 标准章节

每个 SPEC 必须包含以下章节，按固定顺序排列：

```markdown
# <Topic>

## Overview
## <Domain Sections>
## Terminology
## Forbidden
## References
```

### Overview（必需）

回答三个问题：

1. **定义什么？** — 2-3 句话说明范围
2. **不涵盖什么？** — 明确边界，链接到相关 SPEC
3. **相关 SPEC？** — 链接到相关 SPEC

```markdown
## Overview

定义 Token 的命名模式、层级结构和引用规则。

不涵盖 Token 的存储格式和验证规则（见其他相关 SPEC）。
```

> **Why**: 三个问题强制明确边界，防止内容蔓延。读者 30 秒内了解范围和定位。
>
> **Rejected**:
> - 长篇动机说明 — 动机不是范围
> - "Introduction" 标题 — 过于宽泛，容易写成背景故事

### Domain Sections（必需）

SPEC 的核心内容。每个 Domain Section 使用 5 种标准模式之一：

| 模式 | 用途 |
|------|------|
| Pattern A: Schema Definition | 数据格式、字段定义 |
| Pattern B: Rule List | 命名规则、编码规范 |
| Pattern C: Decision Table | 多个相关决策 |
| Pattern D: Interface Contract | API 契约、接口定义 |
| Pattern E: State Machine / Flow | 流程、状态机 |

详见 [13-structure-patterns.md]。

> **Why**: 不同主题需要不同表达方式。标准模式提供一致性，同时保持灵活性。

### Terminology（可选）

仅定义本 SPEC 引入或特化的术语。通用术语链接到 [12-naming.md]。

```markdown
## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Token** | 设计系统的原子单位 | Not "Variable"（Token 是不可变的） |

通用术语见 [12-naming.md §Terminology]。
```

> **Why**: 术语靠近使用它的规则。通用术语集中管理避免重复。

### Forbidden（必需）

列出禁止的模式。格式：`**Don't X**: use Y — reason`

```markdown
## Forbidden

- **Don't use abbreviations in token names**: `color.primary.base` not `clr.pri.base` — abbreviations reduce readability
- **Don't reference tokens by value**: use token name — value may change, name is stable
```

> **Why**: 明确说"不要做什么"比说"应该做什么"更能防止错误。

### References（必需）

链接到相关 SPEC 和外部标准。仅列出读者需要的下一步资源。

```markdown
## References

- [12-naming.md §Terminology] — 术语定义
- [RFC 3339](https://tools.ietf.org/html/rfc3339) — 时间格式标准
```

> **Why**: References 建立 SPEC 之间的导航路径，形成知识网络。

## 内联决策记录

决策记录嵌入在规则旁边，不单独成文件。

### 何时记录决策

用 4 个维度评分（各 1-3 分）：影响范围、持续时间、变更成本、团队影响。

| 总分 | 行动 |
|------|------|
| ≥ 8 | 必须记录（独立 ADR） |
| 5-7 | 建议记录（决策表或内联决策） |
| < 5 | 无需记录 |

> **Why**: 量化标准避免主观判断，确保重要决策被记录。

### 简单决策格式

```markdown
> **Why**: <rationale — 1-3 sentences, factual>
>
> **Rejected**: <alternatives with brief reasons>
```

### 复杂决策格式

多个替代方案时，Rejected 使用列表：

```markdown
> **Why**: Three fixed depths enforce consistent layering across teams.
>
> **Rejected**:
> - User-defined depths — inconsistent layering across teams
> - Flat (no depth) — loses semantic meaning
> - Five-level model — middle levels rarely used
```

### 决策表格式

3+ 个相关决策时，用表格汇总：

```markdown
| Decision | Choice | Why | Rejected |
|----------|--------|-----|----------|
| Separator | `.` | CSS variable aligned | `-` (filename conflict) |
| Case | kebab-case | Cross-platform | camelCase (unclear) |
| Depth | 3 fixed | Balance flexibility | 2 (insufficient), 4+ (complex) |
```

> **Why**: 内联决策将"what"和"why"放在一起，避免分离导致不同步。

## 好坏对比

规则提供好坏对比，使用 ✅ Good 和 ❌ Bad 标记：

```markdown
✅ Good
color.primary.base
spacing.gap.sm

❌ Bad
primaryColor              ← 缺少层级结构
color-primary-base        ← 错误分隔符
```

> **Why**: 好坏对比让 Agent 快速理解边界，避免常见错误。

## 文件大小

| 行数 | 行动 |
|------|------|
| 200-250 | 目标范围 |
| 300+ | 考虑拆分 |
| 400+ | 必须拆分为子 SPEC（同系列内） |

超过 300 行时，按子领域拆分。拆分后更新所有交叉引用。

> **Why**: 200-250 行是单次阅读的舒适范围。超过 400 行说明混合了多个领域。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Overview** | SPEC 的第一章节，回答三个问题 | Not "Introduction" |
| **Domain Section** | SPEC 的核心内容章节 | Not "Body" |
| **Forbidden** | 禁止的模式 + 替代方案 | Not "Anti-pattern" |

## Forbidden

- **Don't include change history**: use git — in-document history becomes stale
- **Don't include implementation tutorials**: use README — SPEC defines contracts, not how-to
- **Don't use "Background & Motivation" sections**: lead with rules, inline brief rationale — motivation essays bury actionable content
- **Don't duplicate definitions from other SPECS**: link to authoritative source — duplication causes drift
- **Don't use vague language**: specify exact pattern — vague rules can't be verified
- **Don't skip Overview's three questions**: always answer what/boundary/related — incomplete Overview causes scope creep
- **Don't mix multiple patterns in one section**: choose one pattern per section — mixing reduces clarity
- **Don't omit Forbidden section**: every SPEC needs Forbidden — omission leads to repeated mistakes

## References

- [00-principles.md] — SPEC 编写核心原则
- [10-organization.md] — SPECS 目录组织规则
- [12-naming.md] — 命名和术语规范
- [13-structure-patterns.md] — Domain Section 的 5 种标准模式
- [20-content-boundaries.md] — 什么内容属于 SPEC
