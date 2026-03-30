---
description: Reviews SPEC documents for completeness, consistency, quality, and technical over-engineering against spec-writing standards and reference implementations. Use when validating specs before implementation, after spec-writing completes a batch, or when specs have been refactored or merged.
name: spec-reviewing
---


# SPEC Reviewing

审查 SPEC 文档的完整性、一致性、质量和技术合理性，确保在进入实施阶段前达到标准。

## Overview

**SPEC review catches gaps AND over-engineering before they become implementation problems.**

本技能覆盖三类审查：

1. **机械审查** — 文件大小、章节结构、引用有效性
2. **语义审查** — 完整性、一致性、可实施性
3. **技术审查** — 过度设计、虚假精确、不必要抽象、与参考实现的偏差

审查标准来自 `spec-writing/guides/` + KISS/DRY/YAGNI/SOLID 原则。

## When to Use

Use when:
- `spec-writing` 完成一批 SPECS 后
- 进入 Code 阶段前（实施前门禁）
- SPEC 重构或合并后
- 提交 SPEC 变更前

When NOT to use:
- 正在编写 SPEC 时（写完再审查）
- 分析 SPEC 与代码的差距（→ `spec-gap-analyzing`）

## Quick Reference

| 检查项 | 标准 | 来源 |
|--------|------|------|
| 章节结构 | Overview → Domain Sections → Terminology → Forbidden → References | `guides/11-structure.md` |
| Overview | 回答：定义什么 / 不涵盖什么 / 相关 SPEC | `guides/11-structure.md` |
| Domain Sections | 使用 5 种标准模式之一 | `guides/13-structure-patterns.md` |
| 内容边界 | 能被违反？是 → 属于 SPEC | `guides/20-content-boundaries.md` |
| 决策记录 | 内联 `> **Why**: ... > **Rejected**: ...` | `guides/11-structure.md` |
| Forbidden | `Don't X: use Y — reason` | `guides/11-structure.md` |
| 文件大小 | 1000 行硬限制 | `guides/11-structure.md` |
| SSOT | 每个概念只在一处定义 | `guides/00-principles.md` |
| 引用有效性 | 所有 `[XX-filename.md]` 链接指向存在的文件 | — |
| 术语一致性 | `SPEC`（全大写），`Agent`（首字母大写） | — |
| **过度设计** | 每个接口/抽象需有参考实现验证 | 技术审查 |
| **虚假精确** | 无实现时不承诺具体数字 | 技术审查 |
| **自我矛盾** | 原则与示例/代码不冲突 | 技术审查 |

## Implementation

### Step 1: 机械检查

对 `SPEC/` 目录做快速扫描：

1. **文件大小** — >1000 行必须拆分
2. **章节结构** — 必须包含 Overview、Forbidden、References
3. **引用有效性** — 所有 `[XX-filename.md]` 链接指向存在的文件
4. **术语一致性** — `SPEC`（非 "Spec"），`Agent`（非 "agent"）

```
🔴 SPEC/03-select.md: 1142 lines — must split
❌ SPEC/01-button.md: Missing Forbidden section
❌ SPEC/02-input.md: Invalid reference [token-api.md]
⚠️ SPEC/02-input.md: Use "SPEC" (not "Spec") on line 45
```

### Step 2: 结构审查

验证每个 SPEC 的章节顺序和内容：

1. **Overview 三问** — 定义什么？不涵盖什么？相关 SPEC？
2. **Domain Sections** — 每个 section 使用 5 种标准模式之一
3. **决策记录** — 重大决策有 `> **Why**` 和 `> **Rejected**`
4. **好坏对比** — 规则包含 ✅ Good 和 ❌ Bad 示例
5. **Forbidden** — 明确的 "Don't X" 列表

```
⚠️ SPEC/01-button.md: Overview missing "not covered" boundary
⚠️ SPEC/01-button.md §Variants: Missing decision record for variant naming
```

### Step 3: 语义审查

检查内容质量和可实施性：

1. **可违反性** — 每条规则能被违反（不能 → 不属于 SPEC）
2. **无模糊语言** — 无 "should"、"might"、"consider"
3. **接口契约完整** — 所有公共 API 有签名、参数、返回类型、错误情况
4. **验收标准可测试** — Given-When-Then 或其他可验证形式
5. **边界情况覆盖** — 错误场景和边界条件已指定

```
❌ SPEC/02-input.md §Validation: "Testing is important" is not violable — remove
⚠️ SPEC/03-select.md: Missing onChange callback parameter types
⚠️ SPEC/01-button.md: Acceptance criterion "button looks good" is not testable
```

### Step 4: 技术过度设计审查

**这是本技能最关键的审查步骤。** 对每个 SPEC 中的技术决策，用以下 7 个检测器逐一评估。

#### 检测器 1：无参考验证的抽象（YAGNI）

**规则**：每个新抽象层（基类、Controller、中间件管线等）必须有至少 1 个参考实现做过类似的事。

**操作**：在 `.references/` 中查找对应参考项目，检查它们是否引入了同样的抽象。

| 信号 | 示例 |
|------|------|
| 定义了基类但参考库没有 | "BehaviorController 强制基类" — Zag/Base UI/React Aria 都没有 behavior 基类 |
| 发明了新管线但参考库按需组合 | "七层标准中间件管道" — Floating UI 用户按需组合，无固定管道 |
| 定义了插件系统但只有 1 个插件 | "可选 FSM 插件：嵌套/并行/历史" — 尚无消费者 |

```
❌ SPEC/03-behavior.md: BehaviorController forced base class — no reference lib uses one
❌ SPEC/10-floating.md: "Seven-layer standard pipeline" — Floating UI has no fixed pipeline concept
⚠️ SPEC/02-state.md: Plugin system for FSM extensions — zero consumers exist
```

#### 检测器 2：第三方 API 重复文档化（DRY）

**规则**：不要在自己的 SPEC 中重新描述第三方库的 API。SPEC 应定义"我们如何封装和使用"，而非库本身怎么工作。

| 信号 | 示例 |
|------|------|
| 规范了库的内部 API 行为 | "Floating UI middleware 执行顺序约束" — 这是 Floating UI 的文档 |
| 复制了库的类型签名 | TanStack Table 的 `ColumnDef<T>` 完整类型 |
| 描述了库的功能矩阵 | "Floating UI 中间件功能对比表" |

```
❌ SPEC/10-floating.md §3: ~120 lines re-documenting Floating UI middleware API
❌ SPEC/12-datagrid.md: Re-specifying TanStack Table's ColumnDef type system
```

#### 检测器 3：虚假精确（YAGNI）

**规则**：没有实现代码时，不应承诺具体的数字指标（体积、性能、行数）。原则性指导优于未经验证的精确数字。

| 信号 | 示例 |
|------|------|
| 精确体积预算但无实现 | "State < 2KB, Behavior < 1.5KB, UI < 1.5KB" |
| 精确性能指标但无基准测试 | "< 16ms 首帧渲染" |
| 代码行数限制但无实际代码 | "UI 层 < 80 行" |
| 插件体积预算但插件不存在 | "嵌套状态 ~1KB, 并行状态 ~1KB" |

```
⚠️ SPEC/01-philosophy.md: Layer-level gzip budgets without implementation data
⚠️ SPEC/02-state.md §2.9: Plugin size budgets for unimplemented plugins
```

#### 检测器 4：自我矛盾

**规则**：SPEC 中的原则/规则不应与同一文档中的示例或代码相矛盾。

**操作**：交叉验证"规则陈述"与"代码示例"。

| 信号 | 示例 |
|------|------|
| 声明"纯函数"但示例包含副作用 | connect 声明纯函数，但案例中 `document.getElementById().focus()` |
| 声明"不持有状态"但示例有本地变量 | behavior 不持有状态，但示例用了 `let isDragging` |
| 声明数字 A 但另一处出现数字 B | "5 包 + subpath exports" vs 列了 8 个包 |

```
❌ SPEC/01-philosophy.md: connect() declared pure but Accordion example calls document.getElementById()
⚠️ SPEC/01-philosophy.md: "5 packages" claim vs 8 packages listed in §4.1
```

#### 检测器 5：过早规范未来特性（YAGNI）

**规则**：V1 不需要的能力不应在 SPEC 中以完整接口形式规范。可以用"Future"附录提及方向，但不应定义接口签名。

| 信号 | 示例 |
|------|------|
| DevTools 调试协议（无调试工具） | `MachineDebugger` 接口 + 时间旅行 `travelTo()` |
| 多平台适配器（只有 Web 端） | RN Behavior connect 完整定义但 RN 尚未启动 |
| 企业级特性（V0.1 阶段） | DataGrid 列冻结 + 行分组 + 虚拟滚动 + 50 万行 |
| 多日历系统（基础组件未完成） | 波斯历、佛历支持 |

```
❌ SPEC/02-state.md §6: ~130 lines DevTools protocol — Zag/Base UI have zero debug infra
❌ SPEC/12-datagrid.md: Full enterprise DataGrid spec at V0.1 — Material Web doesn't have one
⚠️ SPEC/26-calendar.md: Persian/Buddhist calendar systems before basic DatePicker works
```

#### 检测器 6：重复变体（DRY + KISS）

**规则**：同一功能不应有两个几乎相同的实现路径，除非有明确的使用场景差异。

| 信号 | 示例 |
|------|------|
| 两个 Controller 变体功能重叠 | `MachineBehaviorController` vs `SignalMachineBehaviorController` |
| 两种验证模式功能等价 | Zod 直接验证 vs Standard Schema 验证 |
| 两种消费方式未分清场景 | subscribe 模式 vs Signal 模式（项目已选定 Signal） |

```
⚠️ SPEC/02-state.md §7: Two Controller variants with overlapping functionality
```

#### 检测器 7：正向生成 vs 反向提取方向选择

**规则**：如果参考项目全部使用"Code → Metadata"反向提取，而 SPEC 规范了"Metadata → Code"正向生成，需要强有力的理由支撑。

| 信号 | 示例 |
|------|------|
| YAML → Code 但参考库都是 Code → Metadata | 六轨 YAML codegen 但 Stencil/Material Web/Vaadin 都是装饰器 → 提取 |
| 一次性脚手架用完整 IR 管线 | `@generated-once` 文件用 plop/hygen 模板即可 |

```
⚠️ SPEC/17-codegen.md: Forward YAML→Code pipeline, but all references use reverse Code→Metadata
```

### Step 5: 跨 SPEC 一致性

检查 SPECS 之间的一致性：

1. **SSOT 违规** — 同一概念在多个 SPEC 中定义
2. **术语对齐** — 同一概念使用一致的术语
3. **Forbidden 冲突** — 一个 SPEC 禁止的，另一个不应允许
4. **接口对齐** — 共享接口/类型定义一致

```
❌ SSOT: "ComponentStore interface" defined in both philosophy.md §5.1 and 02-state.md §3
⚠️ Terminology: SPEC/01.md uses "variant", SPEC/03.md uses "type" for same concept
```

### Step 6: 生成审查报告并修复

```markdown
# SPEC Review Report
Date: YYYY-MM-DD
Scope: SPEC/

## Summary
- Reviewed: N files
- Issues: X (Critical: N, Warning: N)

## Critical — Technical Over-Engineering
- [ ] SPEC/10-floating.md §3: Re-documents Floating UI middleware API (~120 lines) — delete, keep 1 rule
- [ ] SPEC/02-state.md §6: DevTools protocol without consumers (~130 lines) — move to Future appendix

## Critical — Consistency
- [ ] SPEC/01-philosophy.md: connect() pure function claim contradicts Accordion example
- [ ] SSOT: ComponentStore interface defined in both philosophy.md and 02-state.md

## Warnings
- [ ] SPEC/01-philosophy.md: Layer-level gzip budgets without implementation data
- [ ] SPEC/02-state.md §7: Two Controller variants with overlapping functionality

## Fixes Applied
- SPEC/01-button.md: Added Forbidden section
- SPEC/02-input.md: Removed non-violable statement
```

**简单问题直接修复**（缺少章节、模糊语言），**技术问题标记待讨论**（需确认设计意图），**复杂问题标记待修**（接口缺失、SSOT 违规）。

## Reference Validation Workflow

技术审查时，使用以下流程验证 SPEC 中的设计决策：

```
1. 识别 SPEC 中定义的关键抽象（基类、接口、管线、协议）
2. 在 .references/ 中找到 2+ 个对应的参考项目
3. 检查参考项目是否引入了类似的抽象
4. 如果参考项目没有 → 标记为"无参考验证"
5. 如果参考项目有但更简单 → 标记为"过度复杂"
6. 如果参考项目有且复杂度相当 → 通过
```

**关键对照项目**（按领域）：

| 领域 | 参考项目 | 路径 |
|------|---------|------|
| 状态机 | Zag.js, XState | `.references/headless-behavior/zag/`, `xstate/` |
| Behavior 模式 | Base UI, React Aria | `.references/headless-behavior/base-ui/`, `react-aria/` |
| WC 框架 | Material Web, Spectrum WC | `.references/design-system/material-web/`, `spectrum-web-components/` |
| 浮层定位 | Floating UI (via Zag/Base UI) | `.references/headless-behavior/zag/packages/utilities/popper/` |
| 表格 | TanStack Table, Vaadin Grid | `.references/headless-behavior/tanstack-table/`, `.references/domain-specific/vaadin-grid/` |
| 表单 | TanStack Form, Lion | `.references/headless-behavior/tanstack-form/`, `lion/` |
| Token 系统 | Spectrum CSS, Basecoat | `.references/css-token-system/` |
| 代码生成 | Stencil, @lit-labs/analyzer | `.references/web-component-framework/`, `.references/tooling-infrastructure/` |

## Common Mistakes

**Mistake 1: 只做机械检查，跳过技术审查**
- **Fix:** 技术审查是最高价值的步骤——它发现过度设计、虚假精确和自我矛盾

**Mistake 2: 缺少参考对照就判定"合理"**
- **Fix:** 每个关键抽象都必须有参考实现验证。"听起来合理"≠"经过验证"

**Mistake 3: 将内容量等同于过度设计**
- **Fix:** 过度设计是技术问题（不必要的抽象、虚假精确、无参考验证），不是内容篇幅问题。一个 800 行的 SPEC 如果每条规则都有参考验证，比一个 200 行的 SPEC 包含 3 个无参考抽象要好

**Mistake 4: 审查时加入实现细节**
- **Fix:** 用可违反性测试：不能被违反 → 不属于 SPEC

**Mistake 5: 审查时不看 .references/**
- **Fix:** 技术审查必须打开参考项目源码对照，不能仅凭直觉判断

**Mistake 6: 把第三方库的 API 当成需要审查的"我们的设计"**
- **Fix:** 区分"我们的封装决策"和"库本身的 API 行为"——后者不属于我们的 SPEC

## Forbidden

- **Don't review without reading `spec-writing/guides/`**: 至少读 `00-principles.md`、`11-structure.md`、`20-content-boundaries.md`
- **Don't skip cross-SPEC consistency**: 单文件审查会遗漏 SSOT 违规 — 必须跨 SPEC 检查
- **Don't skip reference validation**: 无参考对照的技术审查等于猜测 — 必须查 `.references/`
- **Don't leave vague findings**: "Needs improvement" 不可执行 — 指明缺什么、在哪里
- **Don't confuse content volume with over-engineering**: 过度设计是技术问题（不必要抽象、虚假精确），不是行数问题
- **Don't audit without fixing**: 审查不修复是浪费 — 简单问题立即修复

## References

- [spec-writing/guides/00-principles.md](../spec-writing/guides/00-principles.md) — 核心原则
- [spec-writing/guides/11-structure.md](../spec-writing/guides/11-structure.md) — SPEC 标准结构
- [spec-writing/guides/13-structure-patterns.md](../spec-writing/guides/13-structure-patterns.md) — 5 种 Domain Section 模式
- [spec-writing/guides/20-content-boundaries.md](../spec-writing/guides/20-content-boundaries.md) — 内容边界规则
- [spec-writing/guides/31-acceptance-criteria.md](../spec-writing/guides/31-acceptance-criteria.md) — 验收标准
