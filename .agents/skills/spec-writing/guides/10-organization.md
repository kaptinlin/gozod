# 目录组织

## Overview

定义 SPECS 目录的组织规则：按领域编号、编号留空隙、ref-* 前缀、文件命名、必需 SPEC。

**不涵盖**：SPEC 内部结构（见 [11-structure.md]）、术语命名（见 [12-naming.md]）、内容边界（见 [20-content-boundaries.md]）。

## 项目目录结构

```
project/
├── SPECS/            # 规范文档（大写 = 权威文档）
├── src/              # 源代码
├── docs/             # 用户文档
└── README.md         # 项目说明
```

> **Why**: 大写目录名突出规范文档的权威性。
> **Rejected**: 全部小写（无法区分规范和代码）。

## 按领域编号

按责任领域编号，不按创建顺序。每个领域占用一个十位数系列。

```
SPECS/
├── 00-principles.md             ← 00-series: Foundation
├── 01-living-documentation.md
├── 10-organization.md           ← 10-series: Structure & Organization
├── 11-structure.md
├── 12-naming.md
├── 20-content-boundaries.md     ← 20-series: Content Standards
├── 21-writing-style.md
├── 31-acceptance-criteria.md    ← 30-series: Decision & Verification
└── 40-spec-templates.md         ← 40-series: Templates & Examples
```

### 典型领域系列

| Series | Domain | Examples |
|--------|--------|---------|
| 00 | Foundation | Principles, naming, terminology |
| 10 | Architecture | Package structure, dependency rules |
| 20 | Data Model | Domain types, schemas, validation |
| 30 | API & Interfaces | HTTP contracts, GraphQL, WebSocket |
| 40 | UI Components | Component specs, props contracts |
| 50 | State Management | Store structure, action patterns |
| 60 | Coding Standards | Language-specific rules, forbidden patterns |
| 70 | Infrastructure | Storage, caching, observability |
| ref-* | References | External standards, project indexes |

> **Why**: 按领域编号使相关 SPEC 聚集在一起。读者可以快速定位主题。

### 必需 SPEC（编码项目）

所有编码项目必须包含：

1. **编码规范**（60-series）：`60-{language}-coding.md`，可选 `61-{framework}-coding.md`
   - 规则分类（CRITICAL > HIGH > MEDIUM）+ 代码示例（✅/❌）+ 禁止事项
   - 基于最新语言/框架特性（Go 1.26+、TypeScript 5.7+、React 19.2+）
   - 详细规范见 [43-coding-standards.md]

2. **设计哲学**（50-series）：`50-philosophy.md` 或 `52-{project}-philosophy.md`
   - 核心价值观、架构原则、权衡取舍、被拒绝方案

非编码项目（文档、配置、设计）不需要编码规范，但建议有设计哲学。

> **Why**: 编码规范确保 Agent 生成符合团队风格的代码。设计哲学确保架构决策一致性。
>
> **Rejected**:
> - 可选编码规范 — 导致代码风格不一致
> - 隐式设计哲学 — 新成员无法理解架构决策
> - 按创建顺序编号 — 相关主题分散，难以导航

## 编号留空隙

系列之间留 10 的间隔（00, 10, 20...），同一系列内不必连续编号。空隙用于未来插入新领域或新 SPEC。

```
00-series: Foundation
10-series: Architecture           ← 系列间隔 10
15-series: [New Domain]           ← 可在空隙插入新系列
20-series: Data Model

10-series 内部:
├── 10-organization.md
├── 11-structure.md
├── 12-naming.md
└── [13-15 可用于未来插入]
```

> **Why**: 预留空隙避免重新编号。重新编号破坏所有交叉引用和外部链接。
>
> **Rejected**: 连续编号 — 插入需要重新编号；更大间隔（100, 200...）— 浪费编号空间。

## ref-* 前缀规则

参考文档使用 `ref-` 前缀，不编号。

```
SPECS/
├── 40-spec-templates.md
├── ref-bdd-practices.md          ← 参考：BDD 实践
└── ref-rfc7807.md                ← 参考：外部标准
```

**是参考文档**：外部标准总结（RFC, W3C, ISO）、理论基础提炼、项目索引
**不是参考文档**：项目自己的规范（→ 编号 SPEC）、实现指南（→ README）、决策记录（→ 内联到 SPEC）

## 文件命名规范

格式：`<number>-<topic>.md` 或 `ref-<topic>.md`

规则：
- **kebab-case**: `31-error-handling.md`，不是 `31_ErrorHandling.md`
- **主题导向**: `20-data-model.md`，不是 `20-user-model.md`（避免具体实体）
- **简洁**: `12-naming.md`，不是 `12-naming-conventions-and-terminology.md`
- **单数形式**: `40-spec-template.md`（例外：复数是主题本质时，如 `00-principles.md`）

> **Why**: kebab-case 是 Markdown 惯例，URL 友好。主题导向避免实现细节泄露。
> **Rejected**: snake_case — 不符合 Markdown 惯例；包含实现细节 — 限制重构。

## 一个领域一个系列

主题跨越两个系列时，归入拥有核心契约的领域。

**示例**：错误处理涉及数据模型（错误类型）和 API 契约（错误响应格式）。核心契约是 API 响应格式 → 放在 30-series。

```
30-series: API & Interfaces
├── 30-api-contracts.md
├── 31-error-handling.md          ← 包含错误类型和响应格式
```

> **Why**: 避免跨系列分散主题。核心契约决定归属。
> **Rejected**: 拆分到多个系列 — 降低内聚性；按实现位置归类 — 实现可能重构，契约更稳定。

## 重新编号策略

避免重新编号 — 破坏交叉引用、git 历史、外部链接。

**仅在以下情况允许**：项目早期（< 5 个 SPEC）、无外部引用、团队一致同意。

**替代方案**：使用空隙编号插入新 SPEC，即使不连续。

```
30-api-contracts.md
31-error-handling.md
32-authentication.md              ← 插入空隙
35-middleware.md                  ← 跳过 33-34 是允许的
```

> **Why**: 重新编号的成本远高于不连续编号的"不美观"。稳定的引用比连续的编号更重要。

## 系列数量限制

保持系列数量在 5-8 个。过多系列说明领域划分过细，合并相关系列。

❌ **坏**：GraphQL 和 REST 各占一个系列 → 合并到 30-series: API & Interfaces
❌ **坏**：Buttons 和 Forms 各占一个系列 → 合并到 40-series: UI Components

> **Why**: 5-8 个系列是认知负担的平衡点。过多系列增加导航成本。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Series** | 一组相关 SPEC 的编号范围（如 10-series） | Not "Category"（series 暗示顺序） |
| **Domain** | 系统的责任领域 | Not "Module"（domain 是逻辑的，module 是物理的） |
| **Gap** | 编号之间的空隙，用于未来插入 | Not "Reserved"（gap 不是预留，是灵活性） |
| **ref-* 文档** | 参考文档，不是规范本身 | Not "Reference SPEC"（不是 SPEC） |
| **核心契约** | 决定主题归属的主要接口或规则 | Not "Primary concern"（契约更具体） |

## Forbidden

- **Don't number by creation order**: 相关主题分散 → 按领域编号
- **Don't use continuous numbering**: 插入时被迫重新编号 → 留空隙（10, 11, 12, 15...）
- **Don't mix domains in one series**: 造成混淆 → 一个系列 = 一个领域
- **Don't number reference documents**: 不是 SPEC → 使用 `ref-` 前缀
- **Don't renumber existing SPECS**: 破坏引用 → 插入空隙或使用下一个可用编号
- **Don't use snake_case or PascalCase**: 不符合 Markdown 惯例 → 使用 kebab-case
- **Don't include implementation details in filename**: 限制重构 → `20-data-model.md` not `20-user-model.md`
- **Don't create more than 8 series**: 导航成本高 → 合并相关领域
- **Don't split related topics across series**: 降低内聚性 → 归入核心契约所属领域

## References

- [11-structure.md] — SPEC 内部结构
- [12-naming.md] — 术语和命名
- [00-principles.md §High Cohesion] — 内聚性原则
- [20-content-boundaries.md] — 内容边界
- [43-coding-standards.md] — 编码规范详细模板
