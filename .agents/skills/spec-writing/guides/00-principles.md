# 核心原则

## Overview

定义编写 SPEC 的六大核心原则，指导 AI Agent 生成正确、可维护的代码，同时支持人类审查意图而非代码。

**不涵盖**：具体实施方法（见 [11-structure.md]）、内容边界（见 [20-content-boundaries.md]）、验证机制（见 [31-acceptance-criteria.md]）。

**相关 SPEC**：所有其他 SPECS 必须遵循这些原则。

## 六大核心原则

### 1. Single Source of Truth (SSOT)

每个概念只在一处定义，其他地方引用。

✅ **好**：在 12-naming.md 定义 Token，其他 SPEC 引用 "见 12-naming.md §Token"
❌ **坏**：在多个 SPEC 重复定义 Token

> **Why**: 避免定义不一致，减少维护成本。
> **Rejected**: 每个 SPEC 独立完整 — 导致定义冲突。

### 2. Current State Only

只记录当前状态，不记录历史。Git 是历史追踪器。

✅ **好**：使用三段式命名 + `> **Why**` 说明依据
❌ **坏**：记录"最初使用扁平命名，后改为三段式"

> **Why**: 文档内历史会腐化，git log 提供完整历史。
> **Rejected**: 变更日志章节 — 与 git 重复且会过时。

### 3. High Cohesion

相关概念放在一起，按领域组织。

✅ **好**：40-architecture-specs.md 包含所有架构规则
❌ **坏**：架构规则分散在多个文件

> **Why**: 高内聚降低查找成本。
> **Rejected**: 按创建时间组织 — 相关概念分散。

### 4. Actionable

包含足够的约束和示例，Agent 能独立实现。

✅ **好**：规则 + 示例 + 验证方法 + 决策依据
❌ **坏**："错误响应应该包含错误信息"（模糊）

**判断标准**：
1. **可重新实现** — 仅凭 SPEC 能否重新实现出功能一致的系统？
2. **稳定性** — 未来 6 个月会改变吗？频繁改变说明是实现细节。
3. **决策性** — 是"决策"（为什么选择这个方案）还是"执行"（具体怎么写代码）？

> **Why**: 模糊规范导致实现不一致。SPEC 定义"选择什么"和"为什么"，而非"如何实现"。
> **Rejected**: 只写高层原则 — Agent 无法推断细节；写实现细节 — 失去 SPEC 的战略价值。

### 5. Concise & Essential

简洁且必要：每句话都有价值，不多不少。

✅ **好**：模式 + 示例 + 禁止项 + 依据（10 行）
❌ **坏**：冗长的背景说明 + 模糊的规则（50 行）

判断标准：
- Agent 能否独立实现？→ 不能则增加约束
- 包含不必要背景？→ 删除
- 过度具体？→ 提升抽象层次

> **Why**: 简洁减少认知负担，必要确保 Agent 有足够信息。
> **Rejected**: 详细背景说明 — 分散注意力；极简主义 — 缺少必要约束。

### 6. Context Minimalism

只给 Agent 完成任务所需的精确信息，不多一个字。

✅ **好**：任务是"修复 API 错误" → 只加载 API specs
❌ **坏**：任务是"修复 API 错误" → 加载所有 SPECS + 所有历史记忆

**任务相关性测试**：
1. **删除测试** — 删除这条信息，Agent 能否完成任务？能 → 删除
2. **替代测试** — 可以在需要时动态获取吗？可以 → 不放入上下文
3. **时效测试** — 在当前任务中会被使用吗？不会 → 删除

> **Why**: 上下文膨胀是 Agent 性能下降的主要原因。对上下文的掌控越好，Agent 表现越好。
> **Rejected**: 全量加载 — 导致上下文膨胀；"以防万一"加载 — 大部分信息用不到。

## SPEC-Driven Development

SPEC 是真相源，代码是产物。人类审查 SPEC（意图、约束、验收标准），不审查代码。

工作流：人类编写 SPEC → Agent 生成代码 → 自动验证 → 人类审查 SPEC 是否解决正确问题

> **Why**: 人类擅长定义意图，Agent 擅长转化为代码。审查 SPEC 比审查代码更高效。
> **Rejected**: 传统代码审查 — 无法跟上 Agent 速度；完全自动化 — Agent 可能误解意图。

## Verification Over Review

用确定性验证替代主观审查，分层验证确保没有单点故障。

**五层验证**：
1. **多选项对比** — 多个 Agent 生成不同实现，选择最佳
2. **确定性护栏** — 测试、类型检查、契约验证
3. **验收标准** — Given-When-Then 场景，Agent 实现，BDD 验证
4. **权限系统** — 限制 Agent 操作范围，特定模式标记人工审查
5. **对抗性验证** — 一个 Agent 实现，另一个验证，第三个尝试破坏

验证标准在代码编写前定义，由人类定义，由机器执行。

> **Why**: 确定性验证可自动化，主观审查无法扩展。分层验证确保即使一层失败，其他层仍能捕获问题。
> **Rejected**: 单一验证层 — 无法捕获所有错误；Agent 自我验证 — 对自己的错误视而不见。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **SPEC** | 定义系统契约、约束、决策的文档 | Not "代码"或"实现" |
| **Agent** | 根据 SPEC 生成代码的 AI 系统 | Not "人类开发者" |
| **SSOT** | Single Source of Truth，唯一真相源 | Not "多处定义" |
| **内联决策记录** | 在 SPEC 中使用 `> **Why**` 和 `> **Rejected**` 记录决策依据 | Not "独立的 ADR 文件" |
| **确定性验证** | 产生明确 pass/fail 结果的验证方法（测试、类型检查等） | Not "主观审查" |

## Forbidden

- **Don't include implementation tutorials**: SPEC 定义决策和约束 → 使用示例放 README
- **Don't record change history**: git 是历史追踪器 → 用 `> **Why**` 说明当前选择依据
- **Don't use vague words**（"应该"、"可能"）: 无法生成确定性代码 → 使用祈使句和明确约束
- **Don't state the obvious**: 浪费上下文空间 → 只记录非显而易见的约束和决策

## References

- [11-structure.md] — SPEC 结构和章节组织
- [20-content-boundaries.md] — 什么内容属于 SPEC
- [31-acceptance-criteria.md] — 验收标准编写
- [52-philosophy-template.md] — Philosophy 文档模板
