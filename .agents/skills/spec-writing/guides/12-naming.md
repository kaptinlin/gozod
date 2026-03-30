# 命名规范

## Overview

定义 SPEC 中的命名规范：术语一致性、术语登记表、大小写规则、语言特定命名。

**不涵盖**：文件命名（见 [10-organization.md §文件命名规范]）、SPEC 结构（见 [11-structure.md]）、内容边界（见 [20-content-boundaries.md]）。

## One Name, One Concept

每个概念使用唯一术语，跨 SPEC、跨语言一致。

✅ **好**：始终用 `scope` 表示 Token 层级
❌ **坏**：混用 `scope`、`tier`、`level` 指代同一概念

> **Why**: 术语一致性降低认知负担。前后端术语一致避免翻译错误。
> **Rejected**: 允许同义词 — 暗示不存在的区别。

## 术语登记表

项目级术语在本 SPEC 的 Terminology 章节登记。SPEC 特定术语在各自 SPEC 中定义。

| 类型 | 定义位置 | 示例 |
|------|---------|------|
| 项目级术语 | 本 SPEC §Terminology | SPEC, Domain, Token, Scope |
| SPEC 特定术语 | 各 SPEC 的 Terminology 章节 | Store, Middleware, Pipeline |
| 外部标准术语 | ref-* 文档 | UUID, RFC 3339, HTTP status |

"Not" 列消除歧义，明确与常见同义词的区别。

> **Why**: 术语登记表是 SSOT。所有 SPEC 链接到这里，避免重复定义。
> **Rejected**: 每个 SPEC 定义所有术语 — 导致分散；单独词汇表文档 — 容易不同步。

## 大小写规则

### 文件和目录

- **kebab-case**: `error-handling.md`, `api-contracts.md`
- **全小写**: `specs/`, `docs/`, `src/`

### React 命名规范

| 元素 | 规则 | 示例 |
|------|------|------|
| 组件名 | PascalCase | `TokenList` |
| 文件名 | kebab-case | `token-list.tsx` |
| Props 类型 | PascalCase | `TokenListProps` |
| Props 属性 | camelCase | `tokens`, `onSelect` |
| Hooks | camelCase, `use` 前缀 | `useTokenFilter` |

### Go 命名规范

| 元素 | 规则 | 示例 |
|------|------|------|
| 包名 | 小写单数 | `token`, `store` |
| 导出类型 | PascalCase | `Token`, `TokenStore` |
| 私有类型 | camelCase | `tokenCache` |
| 导出函数 | PascalCase | `NewTokenStore` |
| 导出错误 | PascalCase, `Err` 前缀 | `ErrNotFound` |
| 接口 | PascalCase, `-er` 结尾或名词 | `Reader`, `TokenStore` |

> **Why**: React kebab-case 文件名 + PascalCase 组件名是社区惯例。Go PascalCase 导出 + camelCase 私有是语言规范。

## 缩写规则

仅使用广泛认可的行业标准缩写（API, HTTP, JSON, UUID, CLI, URL, CSS, HTML, SQL）。

禁止自造缩写：✅ `configuration`, `repository`, `authentication` ❌ `cfg`, `repo`, `auth`

### 首字母缩写词的大小写

| 上下文 | 规则 | 示例 |
|--------|------|------|
| kebab-case | 全小写 | `api-contracts.md` |
| Go PascalCase | 全大写 | `APIClient`, `HTTPServer` |
| React PascalCase | 首字母大写 | `ApiClient`, `HttpServer` |

> **Why**: 标准缩写是行业共识。Go 和 React 对缩写词大小写惯例不同，遵循各自语言规范。

## 禁用同义词列表

以下同义词组在项目中只能选择一个术语使用。

| 同义词组 | 选择 | 理由 |
|---------|------|------|
| Scope / Tier / Level / Layer | `scope`（Token 层级），`layer`（架构层级） | scope 在编程中有明确含义 |
| Manager / Handler / Helper / Utility / Service | 使用领域名词或动词 | 后缀无意义，不传达职责 |
| Config / Configuration / Settings / Options | `config` | 简洁且是行业标准 |
| Error / Exception / Failure | `error`（Go），`Error`（React） | 保持与语言惯例一致 |
| Validate / Check / Verify / Test | `validate`（业务规则），`check`（前置条件） | 区分用途 |
| Get / Fetch / Retrieve / Load / Find | `Get`（同步），`Load`（初始化），`Find`（搜索） | Get 最简洁 |
| Create / Make / New / Build / Add | `New`（构造函数），`Create`（持久化），`Add`（添加到集合） | New 是 Go 惯例 |
| Delete / Remove / Destroy / Drop | `Delete`（持久化），`Remove`（从集合移除） | Delete 是 CRUD 标准 |

### 示例

✅ **好**：`TokenList`（名词）, `useTokenFilter`（动词）, `getToken`, `createToken`, `NewStore`
❌ **坏**：`TokenManager`（无意义后缀）, `fetchToken`（用 get）, `makeToken`（用 create）, `Handler`（模糊）

> **Why**: 无意义后缀不传达职责。统一动词降低认知负担。

## 术语演进

仅在术语与行业标准冲突或造成严重歧义时更改，且团队一致同意。

更改流程：在 Terminology 标记旧术语 deprecated → 添加新定义 → 原子更新所有 SPEC 和代码。

> **Why**: 错误术语比不一致术语更糟，但更改成本高，须谨慎。
> **Rejected**: 禁止更改 — 锁定错误选择；保留旧术语作为别名 — 延续不一致。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Token** | 设计系统的原子单位 | Not "Variable", "Constant" |
| **Scope** | Token 的层级（primitive/semantic/component） | Not "Tier", "Level", "Layer" |
| **Component** | React 组件 | Not "Widget", "Element" |
| **Props** | React 组件属性 | Not "Properties", "Attributes" |
| **Hook** | React Hook 函数 | Not "Helper", "Utility" |
| **Store** | 数据存储接口 | Not "Repository", "DAO" |
| **kebab-case** | 小写加连字符（error-handling） | Not "dash-case", "hyphen-case" |
| **PascalCase** | 首字母大写无分隔符（ErrorHandler） | Not "UpperCamelCase" |
| **camelCase** | 首词小写其余大写（errorCount） | Not "lowerCamelCase" |

## Forbidden

- **Don't use synonyms for the same concept**: 暗示不存在的区别 → 选择一个术语（scope not tier/level）
- **Don't create custom abbreviations**: 增加学习成本 → 仅用标准缩写（API, HTTP, JSON）
- **Don't use meaningless suffixes**: 不传达职责 → `TokenStore` not `TokenManager`
- **Don't mix term definitions across SPECS**: 定义漂移 → 在 Terminology 定义一次，其他处链接
- **Don't reuse terms for different concepts**: 造成歧义 → `context` 不能同时指"上下文对象"和"业务场景"
- **Don't change terms without updating all SPECS**: 部分更新破坏一致性 → 术语更改必须原子化
- **Don't use abbreviations in SPEC prose**: 降低可读性 → 写 `configuration` 不写 `cfg`
- **Don't use PascalCase for React file names**: 不符合 Web 惯例 → 使用 kebab-case
- **Don't add Interface suffix or I prefix to Go interfaces**: 不符合 Go 惯例 → `TokenStore` not `ITokenStore`

## References

- [11-structure.md §Terminology] — Terminology 章节格式
- [10-organization.md §文件命名规范] — 文件命名
- [00-principles.md §Single Source of Truth] — SSOT 原则
- [20-content-boundaries.md] — 内容边界
