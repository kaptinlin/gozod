# SPEC 完整示例

## Overview

提供完整的 SPEC 示例，展示如何应用模板和规范。每个示例包含所有标准章节、内联决策记录、验收标准。

不涵盖：模板结构说明（见 [50-spec-templates.md]）、编写指导（见 [21-writing-style.md]）。

相关 SPEC：[50-spec-templates.md]（模板）、[40-architecture-specs.md]（架构）、[41-api-specs.md]（API）、[42-data-model-specs.md]（数据模型）。

## 示例 1：架构 SPEC

```markdown
# Token 管理系统架构

## Overview

Token 管理系统提供设计 Token 的创建、存储、查询和同步功能。边界：包含 Token CRUD 和版本管理，不包含设计工具集成和 CI/CD 流程。

相关 SPEC：SPECS/60-token-api.md、SPECS/61-token-model.md

## Architecture Views

### Context View

用户：
- **设计师** — 创建和管理 Token
- **开发者** — 查询和使用 Token

外部系统：
- **Figma** — 导入 Token（Figma API）
- **GitHub** — 同步 Token 文件（GitHub API）
- **Slack** — 发送变更通知（Webhook）

数据流：
```
┌──────────┐
│ Designer │
└────┬─────┘
     │ HTTPS
┌────▼────────────┐      ┌─────────┐
│  Web Dashboard  │─────▶│  Figma  │
└────┬────────────┘      └─────────┘
     │ HTTPS/JSON
┌────▼────────────┐      ┌─────────┐
│   API Server    │─────▶│ GitHub  │
└────┬────────────┘      └─────────┘
     │ SQL
┌────▼────────────┐
│   PostgreSQL    │
└─────────────────┘
```

> **Why**: Web Dashboard 和 API Server 分离，支持多客户端（Web、CLI、CI/CD）。

## References

- SPECS/60-token-api.md — Token API 定义
- SPECS/61-token-model.md — Token 数据模型
```

## 示例 2：API SPEC

```markdown
# Token API

## Overview

Token API 提供 Token 的 CRUD 操作。边界：包含创建、查询、更新、删除，不包含批量导入和版本管理。

相关 SPEC：SPECS/61-token-model.md、SPECS/62-token-architecture.md

## Endpoint: GET /api/tokens

### Request

#### 查询参数
- `category` (string, optional) — 枚举：color, spacing, typography
- `limit` (number, optional, default: 100) — 范围：1-1000
- `offset` (number, optional, default: 0) — 分页偏移

#### 请求头
- `Authorization: Bearer <token>` — JWT 认证令牌

### Response (200)

**TypeScript**:
```typescript
interface GetTokensResponse {
  tokens: Token[];
  total: number;
  limit: number;
  offset: number;
}

interface Token {
  id: string;           // UUID
  name: string;         // {category}.{concept}.{variant}
  value: string;
  category: string;     // color|spacing|typography
  description?: string;
  createdAt: string;    // ISO 8601
  updatedAt: string;    // ISO 8601
}
```

**Go**:
```go
type GetTokensResponse struct {
    Tokens []Token `json:"tokens"`
    Total  int     `json:"total"`
    Limit  int     `json:"limit"`
    Offset int     `json:"offset"`
}

type Token struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Value       string    `json:"value"`
    Category    string    `json:"category"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

### Errors

| Code | Condition | Response |
|------|-----------|----------|
| 400 | category 不是枚举值 | `{"error": "invalid_category"}` |
| 400 | limit 超出范围 | `{"error": "invalid_limit"}` |
| 401 | 无效 Authorization | `{"error": "unauthorized"}` |

> **Why**: 查询参数支持灵活过滤和分页。返回 total 支持客户端计算总页数。
> **Rejected**: cursor-based 分页 — offset-based 更简单，数据量 < 100k 足够
```

## 示例 3：数据模型 SPEC

```markdown
# Token 数据模型

## Overview

Token 是设计系统的原子单位。边界：包含 Token 字段定义和验证，不包含 Token 集合和版本管理。

相关 SPEC：SPECS/60-token-api.md、SPECS/62-token-architecture.md

## Schema

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| id | UUID | 是 | 主键 |
| name | string | 是 | 唯一，`^[a-z]+\.[a-z]+\.[a-z]+$` |
| value | string | 是 | 非空，格式随 category |
| category | enum | 是 | color/spacing/typography |
| description | string | 否 | 最大 500 字符 |
| created_at | timestamp | 是 | 不可变 |
| updated_at | timestamp | 是 | 自动更新 |

### Go

```go
type Token struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Value       string    `json:"value"`
    Category    string    `json:"category"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

func (t *Token) Validate() error {
    if t.Name == "" {
        return errors.New("name is required")
    }
    if !slices.Contains(validCategories, t.Category) {
        return errors.New("category must be one of: color, spacing, typography")
    }
    // name prefix must match category
    namePrefix, _, _ := strings.Cut(t.Name, ".")
    if namePrefix != t.Category {
        return errors.New("name prefix must match category field")
    }
    return nil
}
```

> **Why**: 显式 `Validate()` 方法而非 struct tag 验证 — 更清晰、无反射、易测试。
> **Rejected**: `validate` struct tag — 反射性能差，难以实现实体级验证

## Invariants

1. **name 唯一性** — 数据库唯一索引强制约束
2. **name 与 category 一致性** — name 第一段必须与 category 匹配
3. **时间戳单调性** — `updated_at >= created_at`

## Relationships

### Token 与 Version
- **类型**：一对多
- **外键**：Version.token_id → Token.id
- **级联**：删除 Token 时级联删除所有 Version

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Token** | 设计系统的原子单位（name + value + category） | Not "Variable" |
| **Category** | Token 的业务分类枚举 | Not "Type" |
| **Invariant** | 必须始终为真的业务规则 | Not "Constraint"（字段约束） |

## Forbidden

- **Don't modify Token.id**: 不可变唯一标识 → 创建新 Token
- **Don't allow name duplicates**: 导致歧义 → 唯一索引强制
- **Don't store complex objects in value**: Token 是原子单位 → 拆分为多个 Token
- **Don't use struct tags for business validation**: 反射性能差 → 显式 `Validate()` 方法

## References

- SPECS/60-token-api.md — Token API 定义
- SPECS/62-token-architecture.md — Token 架构
- [12-naming.md] — 命名规范
```

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **示例** | 完整的 SPEC 文档，展示模板应用 | Not "模板"（模板是空白框架） |
| **内联决策记录** | `> **Why**` 和 `> **Rejected**` 格式 | Not "独立 ADR 文件" |

## Forbidden

- **Don't skip Overview**: Overview 定义范围和边界 → 始终包含
- **Don't skip decision records**: 决策记录解释"为什么" → 用 `> **Why**` 记录
- **Don't skip Forbidden section**: 明确禁止的模式 → 列出常见错误和替代方案
- **Don't write incomplete examples**: 不完整的示例误导读者 → 提供所有标准章节

## References

- [11-structure.md §内联决策记录] — 决策记录方法
- [31-acceptance-criteria.md] — 验收标准编写
- [40-architecture-specs.md] — 架构 SPEC 规范
- [41-api-specs.md] — API SPEC 规范
- [42-data-model-specs.md] — 数据模型 SPEC 规范
- [50-spec-templates.md] — SPEC 模板库
