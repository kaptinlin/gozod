# 边界案例处理

## Overview

定义模糊内容是否属于 SPEC 的具体判断方法。涵盖质量属性场景、架构决策记录、代码示例、术语表等常见边界案例。不涵盖基础判断标准（见 [20-content-boundaries.md]）、写作风格（见 [21-writing-style.md]）。

## 常见边界案例

### 案例 1: 质量属性场景

"系统必须在特定条件下满足质量要求" — **属于 SPEC**。这是可验证的系统契约，代码可以不满足质量要求。使用 BDD 场景表达：

```markdown
Scenario: 质量属性验证
  Given [系统状态]
  When [触发条件]
  Then [质量要求]
```

### 案例 2: 架构决策记录（ADR）

独立的 ADR 文件 — **不属于 SPEC（作为独立文件）**。决策依据内联到相关规则旁边，作为 `> **Why**` 块。

### 案例 3: 代码示例

展示如何使用 API 的代码片段 — **视情况而定**。

- **属于 SPEC**：展示必须遵守的模式（如错误响应格式）
- **不属于 SPEC**：使用教程（如"如何调用这个 API"）

详细规范见 [24-code-examples.md]。

### 案例 4: 术语表

项目中使用的术语定义 — **属于 SPEC**。术语定义是命名规范的一部分，代码可以使用错误的术语。位置：[12-naming.md §Terminology] 或独立的术语表 SPEC。

## 边界判断表

| 内容 | 能否被违反？ | 结论 |
|------|------------|------|
| "API 必须返回 JSON" | 是（可能返回 XML） | 属于 SPEC |
| "如何配置 JSON 序列化器" | 否（是操作指导） | 不属于 SPEC |
| "禁止使用 `Manager` 后缀" | 是（可能使用） | 属于 SPEC |
| "为什么我们选择 Go" | 否（是背景叙事） | 不属于 SPEC |
| "Given 用户登录 When 访问首页 Then 显示欢迎信息" | 是（系统可能不满足） | 属于 SPEC（验收标准） |

> **Why**: 表格提供快速判断参考。核心标准是"能否被违反"，其他问题是辅助判断。

## Terminology

| Term | Definition |
|------|-----------|
| **质量属性场景** | 定义系统非功能性需求的可验证场景（性能、可用性、安全性等） |
| **ADR** | Architecture Decision Record — SPEC 中使用内联 `> **Why**` 格式替代独立文件 |
| **边界案例** | 难以直接判断是否属于 SPEC 的内容 |

## Forbidden

- **Don't create standalone ADR files**: 决策应内联到规则旁 → 使用 `> **Why**` 块
- **Don't write tutorial-style code examples in SPEC**: 教程属于 README → 只展示必须遵守的模式
- **Don't write vague quality statements**: "系统应该快"无法验证 → 使用 BDD 场景定义可验证的质量属性
- **Don't duplicate terminology definitions**: 违反 SSOT → 在 [12-naming.md §Terminology] 中定义，其他地方引用

## References

- [20-content-boundaries.md] — 内容边界的基础判断标准
- [11-structure.md §内联决策记录] — 如何内联决策依据
- [24-code-examples.md] — 代码示例规范
- [31-acceptance-criteria.md] — 验收标准和 BDD 场景
- [12-naming.md §Terminology] — 术语定义的标准位置
