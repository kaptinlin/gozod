# Philosophy 文档模板

## Overview

提供编写项目 philosophy 或 principles 文档的模板和指南。

定义什么：philosophy 文档的结构模板、写作风格、必需章节。不涵盖：具体项目的哲学内容（由各项目定义）、实现细节。

相关 SPEC：[00-principles.md]（核心原则）、[21-writing-style.md]（写作风格）、[50-spec-templates.md]（SPEC 模板）。

## 核心结构

### 必需章节

#### 1. 核心主张 (Core Thesis)

1-3 句话概括项目哲学，可作为 tagline。核心主张要有立场——暗示你选择了什么、放弃了什么。

**好的核心主张**:
- "Great DX shouldn't come at the expense of great UX" (Svelte) — 揭示编译器策略
- "Clear is better than clever" (Go) — 明确优先级，指导 API 设计
- "Progressive Enhancement" (Alpine) — 暗示渐进引入，无需重写

**弱核心主张**:
- "我们追求高质量的软件" — 没有立场
- "快速、简洁、可靠" — 无优先级，无法指导取舍

#### 2. 原则列表 (Principles)

3-7 个核心原则，每个独立成章。

**命名标准**：简短（2-6 词）、可操作（能指导决策）、易记。

好的原则名称：
- "Write Less Code" (Svelte) — 动词开头，可度量
- "Make the zero value useful" (Go) — 具体技术指导
- "A little copying is better than a little dependency" (Go) — 明确权衡方向
- "Controlled is Cool" (TanStack Form) — 在争议话题上表明立场

弱原则名称：
- "我们重视代码质量和可维护性" — 过于宽泛，无法区分项目
- "使用最佳实践" — 无具体含义

#### 3. 决策依据 (Rationale)

使用 `> **Why**` 和 `> **Rejected**` 块解释"为什么"。

### 可选章节

术语表、应用示例、反模式。

## 原则说明结构

每个原则包含：

1. 一句话定义
2. 详细解释（2-3 段）
3. 正反示例
4. `> **Why**` 和 `> **Rejected**`

**模板**:

```markdown
### [原则名称]

[一句话定义]

[详细解释]

**好的做法**:
[示例]

**不好的做法**:
[反例]

> **Why**: [为什么重要]
> **Rejected**: [拒绝的方案] — [原因]
```

## 哲学类型

### 1. 格言式 (Proverb Style)

简短、诗意、易记。适用成熟项目。参考：Go Proverbs。

**特征**: 每条格言独立可引用。格言不是规则，是共享的智慧。好的格言集覆盖从具体到抽象的层次：技术层 (`"Channels orchestrate; mutexes serialise."`)、设计层 (`"The bigger the interface, the weaker the abstraction."`)、哲学层 (`"Clear is better than clever."`)。

```markdown
### "Don't communicate by sharing memory; share memory by communicating."

Instead of using shared memory protected by locks, pass data between
goroutines using channels. Ownership transfers on send -- no simultaneous
access means no race conditions.

### "A little copying is better than a little dependency."

Sometimes duplicating a small amount of code is better than adding a
dependency. The strconv package implements its own isPrint function
instead of depending on unicode, saving ~150 KB of data tables.

### "Errors are values."

Errors are just values you can program with, not control flow. They can
be wrapped, decorated, stored, or transformed -- consider accumulation,
collection, or decision-making patterns.
```

### 2. 叙事式 (Narrative Style)

从问题到解决方案讲故事。适用新项目。参考：Svelte Philosophy。

**特征**: 先描述行业痛点，再展示项目如何解决。核心是"揭示被忽视的权衡"。

**模式**: 问题 -> 行业现状 -> 洞察 -> 解决方案

```markdown
## Great DX shouldn't come at the expense of great UX

### 问题
Large packages impair user experience -- longer downloads, slower runtime.

### 行业现状
Framework development is a balancing act between DX features and bundle size.

### 洞察
Svelte avoids this trade-off by being a compiler. This decouples DX from UX.

### 解决方案
Developers write code optimally suited for developing. The compiler converts
it at build time to produce an optimal user experience.
```

另一个来自 Svelte 的叙事式原则 "Frameworks Are For Organising Your Mind" 则更短：先点出洞察（框架帮助开发者而非浏览器），再给出结论（编译后消失）。

### 3. 系统式 (Systematic Style)

结构化、全面。适用复杂项目。参考：Alpine Core Principles。

**特征**: 每个原则是平行章节，有统一结构。关键是"正交性"——每个原则覆盖不同维度。Alpine 的维度：语法范式、数据流、架构模式、代码量、迁移策略。

```markdown
### Declarative Syntax

Alpine uses declarative syntax to define behavior directly in HTML markup.
The logic is closely tied to the elements it affects.

> **Why**: 行为和结构在同一处定义，降低心智负担
```

```markdown
### Progressive Enhancement

Alpine can be added to existing HTML without requiring a complete rewrite.

> **Why**: 渐进增强降低采用门槛，允许团队按自己的节奏迁移
```

### 4. 决策式 (Decision Style)

强调权衡和决策依据。面向贡献者。参考：React Design Principles。

**特征**: 每个原则承认权衡的存在，明确选择了什么、放弃了什么。核心是"诚实的权衡"。

```markdown
## Stability

We value API stability. At Facebook, we have more than 20 thousand
components using React. However we think stability in the sense of
"nothing changes" is overrated -- it turns into stagnation. We prefer
"heavily used in production, with a clear migration path when things change."

> **Why**: 稳定不是静止——是有清晰迁移路径的演进
> **Rejected**: 永不破坏 API — 会导致停滞
```

```markdown
## Common Abstraction

We resist adding features that can be implemented in userland. However, if
many components implement a feature in incompatible ways, we bake it into
React. State, lifecycle hooks, event normalization are good examples.

> **Why**: 只在用户空间无法高效解决时才提升为核心抽象
> **Rejected**: 所有功能内置 — 臃肿；所有功能外置 — 碎片化
```

```markdown
## Implementation

We provide elegant APIs where possible but are less concerned with elegant
implementation. We prefer boring code to clever code. Verbose code that is
easy to move around and remove beats prematurely abstracted elegant code.

> **Why**: API 优雅面向用户；实现简单面向维护者
```

## 类型选择指南

| 维度 | 格言式 | 叙事式 | 系统式 | 决策式 |
|------|--------|--------|--------|--------|
| **项目成熟度** | 成熟，大规模验证 | 早期，需要解释存在意义 | 中期，设计空间已明确 | 成熟，有大量贡献者 |
| **主要受众** | 所有开发者 | 潜在采用者 | 使用者 + 贡献者 | 贡献者 + 维护者 |
| **传播性** | 高（可单独引用） | 中（需要完整阅读） | 低（需要系统理解） | 低（面向内部决策） |
| **维护成本** | 低（格言稳定） | 中（故事需要更新） | 高（需保持正交性） | 高（决策需要回顾） |
| **代表项目** | Go Proverbs | Svelte Philosophy | Alpine Principles | React Design Principles |

**混合使用**: 大型项目可混合类型。TanStack Form 混合了决策式（"Controlled is Cool" 表明立场和权衡）和系统式（"Forms need flexibility" 列举多个维度）。

## 原则命名模式

### 好的命名模式

| 模式 | 示例 | 来源 | 效果 |
|------|------|------|------|
| 动词短语 | "Write Less Code" | Svelte | 可操作，指导行为 |
| 祈使句 | "Make the zero value useful" | Go | 像一条指令 |
| 权衡句 | "A little copying is better than a little dependency" | Go | 指明取舍方向 |
| 否定句 | "Don't communicate by sharing memory" | Go | 明确禁止，记忆深刻 |
| 判断句 | "Generics are grim" | TanStack Form | 表明态度 |
| 对仗句 | "Channels orchestrate; mutexes serialise" | Go | 对比清晰，易记 |
| 立场句 | "Controlled is Cool" | TanStack Form | 在争议话题上表明立场 |

### 命名反模式

| 反模式 | 示例 | 问题 |
|--------|------|------|
| 过于宽泛 | "代码质量很重要" | 任何项目都能说，无法区分 |
| 无优先级 | "快速、简洁、可靠" | 并列词无权衡方向 |
| 抽象名词 | "可维护性" | 不可操作 |
| 口号式 | "性能无与伦比" | 营销话术 |

## 文档边界

| 维度 | Philosophy | SPEC | README |
|------|-----------|------|--------|
| **目的** | 指导决策方向 | 定义规则和契约 | 帮助快速上手 |
| **语气** | 引导性（"prefer X over Y"） | 规范性（"MUST / MUST NOT"） | 说明性（"how to"） |
| **受众** | 贡献者 + 设计者 | 实现者 + 审查者 | 使用者 |
| **变更频率** | 低（核心理念稳定） | 中（随实现演进） | 高（随版本更新） |

Philosophy 说"为什么选 X 而不是 Y"。SPEC 说"X 的具体规则是什么"。README 说"如何使用 X"。

## 验证标准

一个好的 philosophy 文档满足：

1. **可引用** — 团队能引用具体原则做决策
2. **可决策** — 面对两个合理方案时，原则能指向其中一个
3. **可传播** — 新成员能在 10 分钟内理解
4. **一致性** — 所有原则相互支撑，无矛盾
5. **可区分** — 能将本项目与同类项目区分

**验证方法**: 取项目中最近 3 个设计决策，用原则论证。找不到对应原则说明文档缺少覆盖；同一决策被两个原则论证为相反结论说明原则间有矛盾。

## 完整模板

```markdown
# [项目名称] Philosophy

## Core Thesis
[1-3 句话。表明立场——你选择了什么，放弃了什么。]

## Principles
### [原则 1: 动词短语或祈使句]
[一句话定义]
[2-3 段详细解释：指导什么决策、什么场景适用、边界条件。]
**好的做法**: [具体正面示例]
**不好的做法**: [具体反面示例]
> **Why**: [为什么这个原则重要]
> **Rejected**: [拒绝的替代方案] — [原因]

### [原则 2-7: 同上结构]

## Terminology
| Term | Definition | Not |
|------|-----------|-----|
| [术语] | [本项目的定义] | Not "[容易混淆的含义]" |

## Forbidden
- **Don't [反模式]**: [为什么禁止] → [应该怎么做]
```

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Philosophy** | 项目的设计理念和核心价值观 | Not "SPEC"（SPEC 定义规则） |
| **Principle** | 从 philosophy 衍生的具体指导原则 | Not "Rule"（指导，不强制） |
| **Proverb** | 简短易记的格言式原则 | Not "详细说明" |
| **Core Thesis** | 一句话概括的核心主张 | Not "完整论述" |
| **Trade-off** | 原则背后的权衡取舍 | Not "缺点"（权衡是有意识的选择） |

## Forbidden

- **Don't write marketing copy**: philosophy 是技术文档 → "优先编译时优化" 而非 "性能无与伦比"
- **Don't over-abstract**: 每个原则必须能指导决策 → "函数不超过 50 行" 而非 "重视代码质量"
- **Don't list obvious principles**: "代码应该可读"不是哲学 → "可读性优于性能，瓶颈时才优化"
- **Don't contradict between principles**: 所有原则形成一致体系 → "显式前提下减少样板" 而非同时主张 "显式优先" 和 "魔法方法"
- **Don't omit examples**: 抽象原则必须有具体示例支撑
- **Don't copy without attribution**: 引用其他项目的原则时标注来源

## References

- [00-principles.md] — SPEC 核心原则
- [11-structure.md &sect;内联决策记录] — 决策记录格式
- [21-writing-style.md] — 写作风格规范
- Go Proverbs (Rob Pike) — 格言式范本
- Svelte Philosophy (Rich Harris) — 叙事式范本
- Alpine Core Principles — 系统式范本
- React Design Principles — 决策式范本
- TanStack Form Philosophy — 混合式范本
