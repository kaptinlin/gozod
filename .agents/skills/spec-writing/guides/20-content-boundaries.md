# 内容边界

## Overview

定义什么内容属于 SPEC，什么不属于。不涵盖 SPEC 结构（见 [11-structure.md §标准章节]）、写作风格（见 [21-writing-style.md]）、Philosophy 文档（见 [52-philosophy-template.md]）。边界案例见 [23-boundary-cases.md]。

- **Philosophy** 定义"为什么这样设计"（设计理念、指导原则）
- **SPEC** 定义"必须遵守的规则"（系统契约、约束、验收标准）

## 判断标准：能否被违反？

SPEC 定义可被违反的规则。如果内容无法被违反，它不属于 SPEC。

### 可被违反的内容（属于 SPEC）

```markdown
✅ Token 命名必须遵循 `{category}.{concept}.{variant}` 模式
   → 可违反：使用 `primary-color` 违反此规则

✅ API 响应必须包含 `id` 字段（UUID v4）
   → 可违反：返回整数 ID 违反此规则

✅ 禁止在 Tier 1 包中导入 Tier 2 包
   → 可违反：添加 `import "tier2/service"` 违反此规则
```

### 不可被违反的内容（不属于 SPEC）

```markdown
❌ "我们使用 React 和 Go 构建系统"
   → 陈述事实，无法违反

❌ "测试很重要"
   → 原则性陈述，无法验证

❌ "运行 `pnpm test` 执行测试"
   → 操作指令，不是约束
```

> **Why**: "能否被违反"是最清晰的边界判断标准。SPEC 定义约束，约束可以被遵守或违反。陈述、教程、历史无法被违反，因此不属于 SPEC。
>
> **Rejected**:
> - "是否重要"作为标准 — 主观，无法一致判断
> - "是否需要文档化"作为标准 — 过于宽泛，包含教程和指南
> - 无明确标准 — 导致 SPEC 膨胀，混入非约束内容

## SPEC 的抽象程度

SPEC 应该在 Schema 层工作，而不是代码层。

### Schema 层 vs 代码层

**Schema 层**（属于 SPEC）：
- 数据模式定义（字段、类型、约束）
- API 契约定义（请求/响应格式、错误码）
- 接口定义（函数签名、参数、返回值）
- 组件 Schema 定义（frontmatter 中的结构化定义）

**代码层**（不属于 SPEC）：
- 具体的算法实现
- 性能优化技巧
- 第三方库的使用方法
- 从 Schema 生成的代码

### 判断标准：可被形式化校验

如果内容可以被机器形式化校验，它属于 SPEC。

✅ **可被形式化校验**（属于 SPEC）：
- JSON Schema 定义 — 可以用 JSON Schema 校验器验证
- API 契约定义 — 可以用 OpenAPI 校验器验证
- 命名规范 — 可以用 Lint 规则验证

❌ **不可被形式化校验**（不属于 SPEC）：
- "这个实现是否优雅" — 主观判断，无法形式化
- "代码是否易于维护" — 主观判断，无法形式化

> **Why**: Schema 层的可校验性使得整个系统的可靠性有了根基。一旦 AI 的工作被约束在 Schema 层，就把 AI 输出的不确定性限制在了一个有完善校验机制的区域。

## 规范文件 vs 执行文件

| 维度 | 规范文件（SPEC） | 执行文件（代码） |
|------|-----------------|-----------------|
| **维护者** | 人 | AI 可自主生成 |
| **内容** | 系统契约、命名规范、架构约束、设计决策、禁止模式、验收标准 | 算法实现、性能优化、第三方库用法 |
| **修改权限** | 只有人可以修改 | AI 自主修改，须符合规范 |

> **Why**: 在工作流规范（CLAUDE.md）中明确写出分层约束，而非依赖 AI 的自觉性。

## 属于 SPEC 的内容

### 系统契约

数据模式、API 契约、接口定义。Schema 定义可被机器形式化校验、可驱动代码生成。

### 命名规范

术语定义、命名模式、一致性规则。

### 架构约束

模块边界、依赖规则、分层架构。

### 设计决策

选择的方案、拒绝的替代方案、决策依据。使用 `> **Why**` 和 `> **Rejected**` 内联到规则旁。

```markdown
## 错误处理策略

使用 sentinel errors + wrapping：

> **Why**: Sentinel errors 支持 `errors.Is()` 判断，wrapping 保留上下文。
>
> **Rejected**:
> - 错误码（error codes）— 需要维护查找表
> - 自定义错误类型 — 简单场景过度设计
```

### 禁止模式

明确不允许的做法和替代方案。格式：`- **Don't X**: reason → use Y instead`

### 验收标准

Given-When-Then 场景（见 [31-acceptance-criteria.md]）。

## 不属于 SPEC 的内容

### 实现细节

具体算法、性能优化、第三方库使用。SPEC 定义"什么"（格式规则），代码实现"如何"（解析逻辑）。

### 操作指南

如何运行、如何部署、如何调试。操作指南是使用说明，不是约束，属于 README。

### 流程定义

代码审查流程、发布流程、协作方式。流程是团队协作方式，不是系统约束，属于 CONTRIBUTING.md。

### 历史记录

变更历史、会议记录、背景故事。历史属于 git，会议记录属于 issue/PR。SPEC 只记录当前约束和决策依据。

### 教程和示例

使用教程、最佳实践指南。SPEC 定义"Token 必须遵循什么格式"，README 说明"如何创建 Token"。

## Terminology

| Term | Definition |
|------|-----------|
| **Schema 层** | 数据模式、API 契约、接口定义等可被形式化校验的抽象层 |
| **代码层** | 具体的算法实现、性能优化等实现细节 |
| **规范文件** | 由人维护，定义行为断言的标准（SPEC） |
| **执行文件** | 由 AI 根据规范文件生成的实现代码 |

## Forbidden

- **Don't write implementation tutorials in SPEC**: 教程属于 README → 在 README 中链接到 SPEC
- **Don't write test cases in SPEC**: 测试用例是实现细节 → 写验收标准（Given-When-Then）
- **Don't write background stories**: 背景叙事易过时 → 用 1-2 句话放入 `> **Why**`
- **Don't create standalone decision log files**: 决策应内联到规则旁 → 使用 `> **Why**` 块
- **Don't maintain change history in SPEC**: Git 提供完整历史 → 使用 `git log --follow`
- **Don't include meeting notes in SPEC**: 会议记录属于 Issue/PR → 在 SPEC 中只记录最终决策

## References

- [11-structure.md §标准章节] — SPEC 的标准章节结构
- [21-writing-style.md] — SPEC 的写作风格规范
- [23-boundary-cases.md] — 边界案例的具体处理方法
- [12-naming.md §Terminology] — 术语定义的标准位置
- [31-acceptance-criteria.md] — 验收标准编写规范
