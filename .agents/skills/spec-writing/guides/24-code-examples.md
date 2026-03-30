# 代码示例规范

## Overview

定义 SPEC 中代码示例的使用规范：何时使用、粒度控制、对比格式。不涵盖内容边界（见 [20-content-boundaries.md]）、写作风格（见 [21-writing-style.md]）。

核心原则：代码示例展示**必须遵守的模式**，而非使用教程。

## 何时使用代码示例

判断标准：

1. **约束 vs 用法？** — 约束（必须遵守的格式）→ SPEC；用法（如何调用）→ README
2. **违反后果？** — 系统不一致、集成失败 → SPEC；用户不会使用 → README
3. **目标读者？** — 实现者（开发者、AI Agent）→ SPEC；使用者 → README

> **Why**: SPEC 定义"系统必须是什么样"，不是"如何使用系统"。

## 示例粒度控制

根据复杂度选择合适的示例层次。

### 层次 1：模式示例（最小化）

只展示必须遵守的模式，不包含完整实现。适用于格式约束、命名规则。

```markdown
## 错误包装

```go
// ✅ 使用 %w 包装错误链
return fmt.Errorf("operation failed: %w", err)

// ❌ 使用 %v 丢失错误链
return fmt.Errorf("operation failed: %v", err)
```
```

### 层次 2：接口示例（中等）

展示接口定义和类型签名，不包含实现逻辑。适用于 API 契约、组件接口、数据模型。

```markdown
## Button 组件

```typescript
interface ButtonProps {
  size: 'sm' | 'md' | 'lg'
  variant: 'solid' | 'outline' | 'ghost'
  disabled?: boolean
}
```

约束：`size` 默认 `'md'`；`disabled` 为 `true` 时忽略点击事件。
```

### 层次 3：参考实现（完整）

仅在必要时提供，用于说明复杂的交互逻辑（状态同步、协调模式）。

### 选择标准

| 判断 | 层次 |
|------|------|
| 一行代码说清楚 | 层次 1（模式） |
| 需要类型定义但不需要实现 | 层次 2（接口） |
| 涉及多步骤交互或非显而易见逻辑 | 层次 3（实现） |

> **Why**: 过少的示例让 AI 无法理解模式，过多的示例变成实现细节。分层控制确保示例恰到好处。

## 示例格式

- **最小化**：只包含展示规则所需的代码，删除无关内容
- **注释关键点**：`✅` 正确做法、`❌` 错误做法、行内注释解释原因
- **指定语言**：代码块必须指定语言（`json`, `typescript`, `go`, `bash`, `yaml`）

```markdown
## Token 命名

模式: `{category}.{concept}.{variant}`

```
✅ color.primary.base
✅ spacing.gap.sm
❌ clr.pri          — 禁止缩写
❌ token.color.primary — 禁止冗余前缀
```
```

## 好坏对比示例

对比示例比单独的"好"或"坏"更有效。标准结构：错误做法 → 正确做法 → 一句话差异原因。

三种格式任选：

1. HTML 注释：`<!-- 不好 -->` / `<!-- 好 -->`
2. 符号：`❌` / `✅`
3. 代码注释：`// ❌ 错误` / `// ✅ 正确`

```markdown
❌ "Error: invalid input"
✅ "Validation failed: email format is invalid"

原因：具体的错误消息帮助用户理解问题和解决方法。
```

> **Why**: 对比示例明确展示了"为什么这样不好"和"怎么做"。单独的示例可能让读者猜测。

## Terminology

| Term | Definition |
|------|-----------|
| **模式示例** | 只展示必须遵守的模式，不含完整实现（层次 1） |
| **接口示例** | 展示接口定义和类型签名，不含实现逻辑（层次 2） |
| **参考实现** | 完整实现，仅用于复杂交互逻辑（层次 3） |
| **对比示例** | 同时展示正确和错误做法 |

## Forbidden

- **Don't write usage tutorials in SPEC**: SPEC 定义"是什么" → 教程放 README
- **Don't provide full implementations (unless necessary)**: 完整实现是实现细节 → 优先使用层次 1 或层次 2
- **Don't include irrelevant code in examples**: 无关代码掩盖核心规则 → 最小化，只保留展示规则所需的代码
- **Don't show only correct examples**: 读者猜测"什么是错的" → 使用对比示例
- **Don't omit language in code blocks**: 无语言标记难以阅读 → 始终用 ```language 指定
- **Don't explain examples verbosely**: 示例应自解释 → 一句话说明差异原因

## References

- [21-writing-style.md] — 写作风格：简洁原则、精确动词
- [20-content-boundaries.md] — 内容边界：什么属于 SPEC
- [11-structure.md §Domain Sections] — Domain Sections 的组织方式
