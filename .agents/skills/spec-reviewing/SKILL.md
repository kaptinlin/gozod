---
description: Reviews SPEC documents for completeness, consistency, quality, and technical over-engineering against spec-writing standards and reference implementations. Use when validating specs before implementation, after spec-writing completes a batch, or when specs have been refactored or merged.
name: spec-reviewing
---


# SPEC Reviewing

审查 SPEC 文档的完整性、一致性、质量和技术合理性。发现问题后**直接修复**，不只是报告。

> 苹果哲学：每个 SPEC 必须回答四个问题——定义什么？什么场景使用？用什么思路解决？是否最优解？无法回答的内容不应存在。

## Overview

**SPEC review catches gaps AND over-engineering before they become implementation problems.**

本技能覆盖三类审查：

1. **机械审查** — 文件大小、章节结构、引用有效性
2. **语义审查** — 完整性、一致性、可实施性
3. **技术审查** — 过度设计、虚假精确、不必要抽象、与参考实现/决策文档的偏差

审查标准来自项目根目录的 `DECISIONS.md`（如存在）+ KISS/DRY/YAGNI 原则 + `.references/` 参考实现。

**核心原则：审查即修复。** 简单问题直接改，复杂问题标记并给出修复方案。不产出只有问题没有答案的报告。

## When to Use

Use when:
- `spec-writing` 完成一批 SPECS 后
- 进入实施阶段前（实施前门禁）
- SPEC 重构或合并后
- 需要验证跨 SPEC 一致性时

When NOT to use:
- 正在编写 SPEC 时（写完再审查）
- 分析 SPEC 与代码的差距（→ `spec-gap-analyzing`）

## Quick Reference

| 检查项 | 标准 |
|--------|------|
| 每个 SPEC 回答四个问题 | 定义什么 / 什么场景 / 什么思路 / 是否最优 |
| 接口契约完整 | 所有公共 API 有签名、参数、返回类型、错误情况 |
| 验收标准可测试 | 每条都能写成自动化测试 |
| 引用有效性 | 所有引用的文件路径存在 |
| 决策对齐 | 所有设计选择与 DECISIONS.md 一致（如存在） |
| **过度设计** | 每个抽象需有参考实现验证 |
| **虚假精确** | 无实现时不承诺具体数字 |
| **自我矛盾** | 原则与示例/代码不冲突 |
| **使用体验** | 零配置入口、渐进披露、一步完成、命名直觉、语义方法、概念唯一 |
| **API spec 完整性** | 是否存在一份从消费者角度定义公共 API 的规范文档，而不是只有实现结构或内部模块拆分 |
| 跨 SPEC 一致性 | 共享类型定义一致、术语对齐、无 SSOT 违规 |

## Implementation

### Step 1: 收集上下文

在审查 SPECS 前，先读取项目上下文：

1. **DECISIONS.md**（如存在）— 所有设计决策及其 ID（如 D01–D17）
2. **IMPROVE.md**（如存在）— 设计审视和改进建议
3. **SPECS/01-requirements.md**（如存在）— 需求基线
4. **ANALYSIS.md**（如存在）— 规范阶段分析

这些文档是审查的**对照基准**。SPEC 中的设计选择必须与决策文档一致。

### Step 2: 机械检查

对 `SPECS/` 目录做快速扫描：

1. **文件大小** — >1000 行必须拆分
2. **引用有效性** — 所有引用的文件路径指向存在的文件
3. **决策 ID 引用** — 所有 `[DXX]` 引用指向 DECISIONS.md 中存在的决策

```
🔴 SPECS/04-unified-format.md: 1142 lines — must split
❌ SPECS/03-text-diffing.md: References D99 — no such decision in DECISIONS.md
⚠️ SPECS/05-hunk-operations.md: References .research/R09-xxx.md — file doesn't exist
```

### Step 3: 四个问题审查

对每个 SPEC，验证它是否清晰回答了四个问题：

| 问题 | 在 SPEC 中的体现 | 缺失时的症状 |
|------|-----------------|-------------|
| **定义什么？** | Overview/Scope 段落 | 读完不知道这个模块的边界在哪 |
| **什么场景使用？** | Consumer/Use Case 描述 | 不知道谁会调用这些 API |
| **什么思路解决？** | 算法描述、类型设计、关键决策 | 只有接口签名没有设计理由 |
| **是否最优解？** | 决策记录、备选方案对比、参考实现引用 | 选了一种方案但不知道为什么不选其他的 |

```
❌ SPECS/06-hashline-encoding.md: Missing "什么场景使用" — no consumer mapping
⚠️ SPECS/03-text-diffing.md: "什么思路解决" lacks algorithm description — only has function signatures
```

### Step 4: 技术过度设计审查

对每个 SPEC 中的技术决策，用以下检测器评估。

#### 检测器 1：无参考验证的抽象（YAGNI）

**规则**：每个新抽象层必须有至少 1 个参考实现做过类似的事。

**操作**：在 `.references/` 中查找对应参考项目，检查它们是否引入了同样的抽象。

| 信号 | 示例 |
|------|------|
| 定义了基类但参考库没有 | "DiffEngine 基类" — 参考实现全部是独立函数 |
| 发明了新管线但参考库按需组合 | "五阶段 diff 管道" — znkr-diff 直接调用 |
| 定义了插件系统但无消费者 | "可选 diff 算法插件" — 尚无插件 |

```
❌ SPECS/02-root-types.md: DiffEngine interface — no reference lib uses one
⚠️ SPECS/08-structured.md: Plugin system for format parsers — only JSON exists
```

#### 检测器 2：虚假精确（YAGNI）

**规则**：没有实现代码时，不应承诺具体的数字指标（体积、性能、行数）。原则性指导优于未经验证的精确数字。

| 信号 | 示例 |
|------|------|
| 精确性能指标但无基准测试 | "< 16ms for 5K lines" 但无 benchmark |
| 代码行数限制但无实际代码 | "internal/myers.go < 200 lines" |
| 内存预算但无 profiling | "O(N) memory, peak < 50MB" |

```
⚠️ SPECS/03-text-diffing.md: "< 50ms for 10K lines" — acceptable if from reference benchmark
⚠️ SPECS/08-structured.md: "O(K×D) time" — verify after implementation
```

#### 检测器 3：自我矛盾

**规则**：SPEC 中的原则/规则不应与同一文档或其他 SPEC 中的内容矛盾。

| 信号 | 示例 |
|------|------|
| 声明"stdlib only"但依赖外部库 | structured/ 声明零依赖但需要 json/v2 |
| 声明类型 A 但另一处定义类型 B | Edit.OldPos vs Edit.OldLine |
| 声明不实现但另一处要求实现 | "no Decode" vs acceptance criteria mentions round-trip |

```
❌ SPECS/02-root-types.md: Edit uses OldPos — DECISIONS.md D02 says OldLine
❌ SPECS/06-hashline.md: Acceptance criteria mentions Decode but D09 says encode-only
```

#### 检测器 4：过早规范未来特性（YAGNI）

**规则**：当前阶段不需要的能力不应以完整接口形式规范。可以用"Deferred"段落提及方向，但不应定义接口签名。

| 信号 | 示例 |
|------|------|
| Phase 4 功能出现在 Phase 1 SPEC 中 | DiffRunes 完整签名在 text/ Phase 1 SPEC |
| YAML 完整接口在 JSON-only 阶段 | DiffYAML 签名在 structured/ Phase 3 SPEC |
| 三种模式全部规范但 Phase 1 只实现一种 | Minimal/Fast 选项在 text/ Phase 1 SPEC |

```
⚠️ SPECS/03-text-diffing.md: DiffRunes full signature — should be in Deferred section only
❌ SPECS/08-structured.md: DiffYAML interface defined — YAML ecosystem not ready
```

#### 检测器 5：使用体验缺陷

**规则**：SPEC 定义的 API 必须从使用者角度审视，而非仅从实现者角度。好的规范让 90% 的场景用一行代码完成。

**额外要求**：对于有公共库/API 的项目，SPEC 体系中必须存在一份独立的消费者 API 设计文档或等效 API spec。它描述的是 public API 应该如何被消费，而不是代码内部如何实现。文档必须优先回答：默认入口是什么、默认构造方式是什么、常见任务最短路径是什么、哪些内部层级/平台细节不应该泄漏给普通用户。

**六个审查维度**：

| 维度 | 审查要点 |
|------|---------|
| **零配置入口** | 是否存在无需配置即可使用的顶层入口？用户写的第一行代码是创建对象还是拼装依赖？ |
| **渐进披露** | 核心类型是否按使用频率分层？高频类型是否被低频类型淹没？ |
| **一步完成** | 完成一个常见任务需要几步？跨包调用、手动组装、中间转换是否可以合并？ |
| **命名直觉** | 类型/常量命名是否利用了包前缀提供的命名空间？是否存在冗长前缀叠加？ |
| **语义方法** | 核心数据结构是否有便捷方法？还是强迫调用者做枚举值比较？ |
| **概念唯一** | 同一语义是否有两套表达机制？如果有，两者的职责边界是否明确？ |

**操作**：对 SPEC 中定义的每个公共 API，模拟用户最常见的使用路径，检查是否存在上述缺陷。

如果项目缺少这类消费者 API spec，审查时必须将其标记为结构性缺口：

- ❌ 只有内部模块/包拆分，没有消费者 API 入口规范
- ❌ 只有实现级接口签名，没有 consumer journey
- ❌ API 文档围绕 backend/platform/driver/store/plugin 等内部实现组织，而不是围绕用户任务组织
- ✅ 有单独 API spec，明确默认入口、默认构造、常见用法、渐进披露和禁止泄漏的内部复杂度

#### 检测器 6：决策文档偏差

**规则**：如果项目有 DECISIONS.md，SPEC 中的每个设计选择必须与对应决策一致。

**操作**：逐个比对 SPEC 中的关键设计点与 DECISIONS.md 中的决策。

| 信号 | 示例 |
|------|------|
| SPEC 选了备选方案而非推荐方案 | SPEC 用字节偏移但 D02 选行号 |
| SPEC 引入了决策中否决的依赖 | SPEC 依赖 jsondiff 但 D07 说不依赖 |
| SPEC 遗漏了决策中要求的内容 | D13 要求预处理但 SPEC 未提及 |

```
❌ SPECS/04-unified.md: Uses io.Reader entry — D05 says Parse([]byte) only
❌ SPECS/08-structured.md: Depends on jsondiff — D07 says self-contained
⚠️ SPECS/03-text-diffing.md: Missing preprocessing section — D13 requires it
```

### Step 5: 跨 SPEC 一致性

检查 SPECS 之间的一致性：

1. **共享类型对齐** — Edit/EditScript/OpType 在所有 SPEC 中定义一致
2. **术语对齐** — 同一概念使用一致的术语（如 "hunk" vs "chunk"）
3. **依赖边界对齐** — 核心包 stdlib only，扩展包允许外部依赖
4. **Phase 边界对齐** — 每个 SPEC 明确属于哪个 Phase，不跨 Phase 引用未实现的能力
5. **API 签名对齐** — 一个 SPEC 中定义的函数参数/返回类型与消费方 SPEC 中的使用一致

```
❌ SSOT: Edit type defined differently in SPECS/02 (OldPos) and SPECS/05 (OldLine)
⚠️ Terminology: SPECS/03 uses "edit script", SPECS/05 uses "edit sequence" for same concept
❌ Phase boundary: SPECS/05-hunk uses structured.Change type but structured/ is Phase 3
```

### Step 6: 修复并生成报告

**简单问题直接修复**（术语不一致、遗漏决策引用、类型名不对齐）。
**复杂问题标记并给出修复方案**（设计偏差、跨 SPEC 矛盾）。

```markdown
# SPEC Review Report
Date: YYYY-MM-DD
Scope: SPECS/02–SPECS/08

## Summary
- Reviewed: N files
- Fixed: X issues
- Remaining: Y issues (need discussion)

## Fixes Applied
- SPECS/02-root-types.md: Edit.OldPos → Edit.OldLine (align with D02)
- SPECS/06-hashline.md: Removed Decode acceptance criteria (align with D09)
- SPECS/03-text-diffing.md: Added preprocessing section (align with D13)
- All SPECS: Unified terminology "EditScript" (was "edit sequence" in 3 places)

## Remaining Issues
- [ ] SPECS/08-structured.md: DiffYAML interface defined but YAML ecosystem not ready — recommend moving to Deferred section
- [ ] SPECS/03-text-diffing.md: Performance target "< 50ms" needs benchmark validation post-implementation
```

## Reference Validation Workflow

技术审查时，使用以下流程验证 SPEC 中的设计决策：

```
1. 识别 SPEC 中定义的关键抽象（类型、接口、算法选择）
2. 在 DECISIONS.md 中找到对应决策 → 检查是否一致
3. 在 .references/ 中找到 2+ 个对应的参考项目
4. 检查参考项目是否引入了类似的抽象
5. 如果参考项目没有 → 标记为"无参考验证"
6. 如果参考项目有但更简单 → 标记为"过度复杂"
7. 如果参考项目有且复杂度相当 → 通过
```

## Common Mistakes

**Mistake 1: 只做机械检查，跳过技术审查**
- **Fix:** 技术审查是最高价值的步骤——它发现过度设计、虚假精确和决策偏差

**Mistake 2: 审查后只输出报告，不修复**
- **Fix:** 审查即修复。能改的立即改。不能改的给出明确修复方案

**Mistake 3: 缺少决策文档对照**
- **Fix:** 如果项目有 DECISIONS.md，每个设计选择都必须与对应决策比对。"SPEC 这么写了"≠"设计意图如此"

**Mistake 4: 忽略跨 SPEC 一致性**
- **Fix:** 单文件审查会遗漏共享类型不一致、术语分裂、Phase 边界越界

**Mistake 5: 把内容量等同于过度设计**
- **Fix:** 过度设计是技术问题（不必要抽象、虚假精确、无参考验证），不是行数问题

## Forbidden

- **Don't review without reading context documents**: 先读 DECISIONS.md、IMPROVE.md、SPECS/01-requirements.md
- **Don't skip cross-SPEC consistency**: 单文件审查会遗漏 SSOT 违规 — 必须跨 SPEC 检查
- **Don't skip reference validation**: 无参考对照的技术审查等于猜测 — 必须查 `.references/`
- **Don't leave vague findings**: "Needs improvement" 不可执行 — 指明缺什么、在哪里、怎么改
- **Don't audit without fixing**: 审查不修复是浪费 — 简单问题立即修复
- **Don't accept decision drift**: SPEC 与 DECISIONS.md 不一致时，以 DECISIONS.md 为准修复 SPEC
- **Don't ignore user perspective**: 实现者觉得合理的 API 不等于使用者觉得好用 — 必须模拟使用路径
