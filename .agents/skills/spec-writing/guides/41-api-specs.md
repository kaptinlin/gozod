# API SPEC 规范

## Overview

定义如何编写 API 相关的 SPEC。API SPEC 记录接口契约：请求格式、响应格式、错误处理、版本控制。

不涵盖 API 实现细节（代码注释）、API 使用教程（README）、API 测试方法（测试代码）。

相关 SPEC：[11-structure.md]（SPEC 结构）、[42-data-model-specs.md]（数据模型定义）、[31-acceptance-criteria.md]（验收标准）。

## API 契约定义

### 端点规范

每个端点定义：路径模式、HTTP 方法、请求格式、响应格式、错误响应。

**示例**:
```markdown
## 用户查询

### 端点
`GET /api/v1/tokens/{id}`

### 路径参数
- `id` (string, required) — 用户 ID，UUID 格式

### 响应 (200 OK)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "color.primary.base",
  "value": "#0066CC",
  "category": "color",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 错误响应
- `404 Not Found` — 用户不存在
- `400 Bad Request` — ID 格式无效

> **Why**: 明确的契约确保客户端和服务端的一致性。UUID 格式防止枚举攻击。
```

### 请求格式

定义请求体结构、字段类型和约束、必填/可选字段、验证规则。

**示例**:
```markdown
## 创建用户

### 端点
`POST /api/v1/tokens`

### 请求体
```json
{
  "name": "color.primary.base",
  "value": "#0066CC",
  "category": "color"
}
```

### 字段规范

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| name | string | 是 | Token 名称格式，唯一 |
| value | string | 是 | 非空 |
| category | string | 是 | 枚举：color, spacing, typography |

### 验证规则
- Name 必须匹配 Token 命名模式
- Category 必须是枚举值之一
- Value 格式必须符合 category 要求

> **Why**: 明确的验证规则防止无效数据进入系统。在 API 层验证，
> 而非依赖客户端验证。
```

### 响应格式

定义成功响应结构、HTTP 状态码、响应头（如需要）。

**示例**:
```markdown
## 响应格式

### 成功响应 (201 Created)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "color.primary.base",
  "value": "#0066CC",
  "category": "color",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 响应头
- `Location: /api/v1/tokens/{id}` — 新创建资源的 URL
- `Content-Type: application/json`

> **Why**: 201 状态码明确表示资源已创建。Location 头允许客户端
> 直接访问新资源，无需解析响应体。
```

## 错误处理规范

### RFC 7807 Problem Details

所有错误响应遵循 RFC 7807 格式。

**标准格式**:
```markdown
## 错误响应格式

所有 API 错误返回 RFC 7807 Problem Details：

```json
{
  "type": "https://api.example.com/errors/validation-failed",
  "title": "Validation Failed",
  "status": 400,
  "detail": "Email format is invalid",
  "instance": "/api/v1/tokens",
  "errors": [
    {
      "field": "email",
      "message": "Must be a valid email address"
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string (URI) | 是 | 错误类型的唯一标识符 |
| title | string | 是 | 人类可读的错误摘要 |
| status | integer | 是 | HTTP 状态码 |
| detail | string | 否 | 具体的错误描述 |
| instance | string (URI) | 否 | 发生错误的请求路径 |
| errors | array | 否 | 字段级错误详情（验证错误时使用） |

> **Why**: RFC 7807 是标准格式，工具支持好。结构化的错误信息
> 便于客户端解析和展示。
>
> **Rejected**:
> - 自定义错误格式 — 增加客户端解析复杂度
> - 纯文本错误消息 — 难以程序化处理
> - 错误码（如 E1001） — 需要额外的码表维护
```

### 错误类型定义

预定义错误类型表：

**示例**:
```markdown
## 错误类型

| Type URI | Status | Title | 使用场景 |
|----------|--------|-------|---------|
| `/errors/not-found` | 404 | Resource Not Found | 资源不存在 |
| `/errors/validation-failed` | 400 | Validation Failed | 请求验证失败 |
| `/errors/unauthorized` | 401 | Unauthorized | 未认证 |
| `/errors/forbidden` | 403 | Forbidden | 无权限 |
| `/errors/conflict` | 409 | Conflict | 资源冲突（如重复创建） |
| `/errors/internal` | 500 | Internal Server Error | 服务器内部错误 |

> **Why**: 预定义的错误类型确保一致性。客户端可以根据 type URI
> 实现特定的错误处理逻辑。
```

### 验证错误详情

验证失败时，`errors` 字段包含字段级错误：

```json
{
  "field": "email",
  "message": "Must be a valid email address",
  "code": "invalid_format"
}
```

- `field` (string) — 字段路径（支持嵌套，如 `address.city`）
- `message` (string) — 人类可读的错误消息
- `code` (string) — 机器可读的错误代码，支持国际化

## 版本控制策略

### URL 版本控制

**规范**:
```markdown
## API 版本控制

使用 URL 路径版本控制：`/api/v{major}/...`

### 版本格式
- `v1`, `v2`, `v3` — 主版本号
- 不包含次版本号（次版本向后兼容，无需新路径）

### 示例
- `/api/v1/tokens` — 版本 1
- `/api/v2/tokens` — 版本 2（不兼容的变更）

### 版本生命周期
- 新版本发布后，旧版本维护 12 个月
- 废弃通知提前 6 个月
- 响应头包含 `Sunset` 头标注废弃日期

> **Why**: URL 版本控制简单明确，易于路由和缓存。主版本号足够，
> 次版本通过向后兼容的方式演进。
>
> **Rejected**:
> - Header 版本控制（`Accept: application/vnd.api.v1+json`） — 难以测试和缓存
> - 查询参数版本控制（`?version=1`） — 容易被忽略
> - 无版本控制 — 无法引入不兼容变更
```

### 兼容性判断

| 变更类型 | 兼容性 | 需要新版本 |
|----------|--------|-----------|
| 添加新端点 | 兼容 | 否 |
| 添加可选请求字段 | 兼容 | 否 |
| 添加响应字段 | 兼容 | 否 |
| 放宽验证（如字段长度增加） | 兼容 | 否 |
| 添加新错误类型 | 兼容 | 否 |
| 删除/重命名端点或字段 | **不兼容** | 是 |
| 改变字段类型 | **不兼容** | 是 |
| 收紧验证 | **不兼容** | 是 |
| 改变语义（同输入不同输出） | **不兼容** | 是 |

> **Why**: 客户端应容忍未知字段（Postel's Law）。不兼容变更通过新版本隔离。

## 认证授权规范

API SPEC 中认证授权部分必须定义：

**认证**:
- 认证方式和请求头格式
- Token 类型和有效期
- 未认证响应（401）

**授权**:
- 授权模型（RBAC/ABAC/ACL）
- 角色和权限映射表
- 无权限响应（403）

所有错误响应遵循 RFC 7807 格式。

> **Why**: 认证授权是 API 安全的基础。明确定义避免实现歧义。

## Terminology

| Term | Definition |
|------|-----------|
| **API 契约** | 客户端与服务端之间的接口约定（路径、请求、响应、错误） |
| **RFC 7807** | Problem Details for HTTP APIs，标准化错误响应格式 |
| **向后兼容** | 变更不破坏现有客户端的正常使用 |

## Forbidden

- **Don't 包含 API 实现代码**: SPEC 定义契约，不是实现 → 只定义请求/响应格式
- **Don't 包含客户端使用示例**: 使用教程属于 README 或 API 文档 → 只展示请求/响应格式
- **Don't 包含性能测试结果**: 测试结果属于测试报告 → 定义性能要求（场景）
- **Don't 使用模糊的错误消息**: "Error occurred" 无法帮助调试 → 具体错误消息，包含上下文和解决建议
- **Don't 在响应中暴露内部实现**: 如数据库字段名、堆栈跟踪 → 使用业务语言，隐藏实现细节
- **Don't 在 URL 中使用动词**: RESTful API 使用 HTTP 方法表示动作 → GET 查询，POST 创建

## References

- [11-structure.md] — SPEC 结构
- [20-content-boundaries.md] — 内容边界
- [40-architecture-specs.md] — 架构约束
- [42-data-model-specs.md] — 数据模型定义
- [31-acceptance-criteria.md] — BDD 场景
- RFC 7807 — Problem Details for HTTP APIs
- OpenAPI Specification — API 定义标准
