# SPEC 结构模式

## Overview

定义 SPEC Domain Section 的 5 种标准模式：如何选择、如何使用。

不涵盖章节顺序（见 [11-structure.md]）、内容边界（见 [20-content-boundaries.md]）、写作风格（见 [21-writing-style.md]）。

## Pattern A: Schema Definition

用于数据格式、字段定义、API 响应。关键元素：字段表格（Field, Type, Required, Constraints）+ Invariants（跨字段不变量）。

```markdown
## Token Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | yes | 格式: `{category}.{concept}.{variant}` |
| `value` | string | yes | 非空 |
| `category` | string | yes | 枚举: `color`, `spacing`, `typography` |

### Invariants

- Token.name 必须唯一
- Token.category 必须与 name 的第一段匹配
```

> **Why**: 表格提供快速概览，Invariants 定义不变量，Agent 可以生成验证代码。
>
> **Rejected**: 纯文字描述（难以扫描），JSON Schema（过于技术化）。

## Pattern B: Rule List

用于命名规则、编码规范、执行顺序。关键元素：模式定义 + 示例 + 约束列表 + ✅/❌ 对比。

```markdown
## Token 命名规则

模式：`{category}.{concept}.{variant}`

示例：
- `color.primary.base` — 主色基础值
- `spacing.gap.sm` — 小间距

约束：
- 使用 kebab-case，不使用 camelCase
- 每段 2-15 字符
- 不使用缩写（`clr` → `color`）
```

> **Why**: 规则列表适合线性、独立的约束。示例让规则具体化。
>
> **Rejected**: 表格格式（规则不需要多列对比），纯示例（缺少明确规则）。

## Pattern C: Decision Table

用于 3+ 个相关决策的对比。关键列：决策点、选择、理由、拒绝方案。

```markdown
## 命名决策

| 决策点 | 选择 | 理由 | 拒绝方案 |
|--------|------|------|----------|
| 分隔符 | `.` | 与 CSS 变量一致 | `-`（与文件名冲突）, `_`（不够清晰） |
| 大小写 | kebab-case | 跨平台兼容 | camelCase（不够清晰） |
| 层级数 | 3 层固定 | 平衡灵活性和一致性 | 2 层（不够）, 4+ 层（过于复杂） |
```

> **Why**: 决策表避免重复的"Why/Rejected"块，提供快速概览。
>
> **Rejected**: 每个决策单独一节（过于冗长），纯文字列表（难以对比）。

## Pattern D: Interface Contract

用于 API 边界、类型接口、模块间契约。关键元素：Request/Response 双语言定义 + Errors 表格。

```markdown
## Token API

### Request

\`\`\`typescript
interface GetTokensRequest {
  category?: string;  // 可选过滤
  limit?: number;     // 默认 100
}
\`\`\`

\`\`\`go
type GetTokensRequest struct {
    Category string `json:"category,omitempty"`
    Limit    int    `json:"limit,omitempty"` // 默认 100
}
\`\`\`

### Errors

| Code | Condition | Response |
|------|-----------|----------|
| 400 | Invalid category | `{"error": "invalid_category"}` |
| 500 | Database error | `{"error": "internal_error"}` |
```

> **Why**: 双语言定义确保前后端契约一致，Agent 可以生成类型安全的代码。
>
> **Rejected**: 仅后端定义（前端需手动同步），OpenAPI（过于冗长）。

## Pattern E: State Machine / Flow

用于流程、管道、生命周期。关键元素：ASCII 流程图 + 状态转换表。

```markdown
## Token 生命周期

Draft → Review → Approved → Published → Archived

| State | Allowed Transitions | Trigger |
|-------|-------------------|---------|
| Draft | Review, Rejected | 提交审查 / 拒绝 |
| Review | Approved, Rejected | 审查通过 / 拒绝 |
| Approved | Published, Rejected | 发布 / 拒绝 |
| Published | Archived | 归档 |
```

> **Why**: ASCII 图在 Markdown 中直接可见，状态表定义转换规则，Agent 可以生成状态机代码。
>
> **Rejected**: Mermaid 图（需要渲染工具），纯文字描述（难以理解流程）。

## 模式选择指南

| 内容类型 | 推荐模式 |
|---------|---------|
| 数据结构、字段定义 | Pattern A: Schema Definition |
| 命名规则、编码规范 | Pattern B: Rule List |
| 多个相关决策 | Pattern C: Decision Table |
| API 契约、接口定义 | Pattern D: Interface Contract |
| 流程、状态机 | Pattern E: State Machine / Flow |

一个 SPEC 可以包含多个模式，但每个 Domain Section 只使用一种模式。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Pattern** | Domain Section 的标准结构模式 | Not "Template"（Pattern 是结构，不是填空） |
| **Invariant** | 跨字段的不变量约束 | Not "Constraint"（Constraint 是单字段） |

## Forbidden

- **Don't mix multiple patterns in one section**: choose one pattern per section — mixing reduces clarity
- **Don't use Mermaid or PlantUML**: use ASCII diagrams — external tools add rendering dependency
- **Don't use JSON Schema in SPEC**: use table format — JSON Schema is too technical for human review
- **Don't create custom patterns**: use one of the 5 standard patterns — custom patterns reduce consistency
- **Don't use decision table for single decision**: use inline Why/Rejected — decision table is for 3+ related decisions
- **Don't omit examples in Rule List**: always provide ✅/❌ examples — examples make rules concrete

## References

- [11-structure.md] — 标准章节、顺序、内联决策记录
- [20-content-boundaries.md] — 内容边界判断
- [21-writing-style.md] — 写作风格规范
