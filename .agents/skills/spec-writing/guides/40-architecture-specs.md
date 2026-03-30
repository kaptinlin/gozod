# 架构 SPEC 规范

## Overview

定义如何编写架构相关的 SPEC。架构 SPEC 记录系统的关键决策、质量属性、约束和边界。

不涵盖详细设计（组件内部实现）、部署配置（基础设施代码）、架构图工具使用。

相关 SPEC：[11-structure.md]（SPEC 结构）、[11-structure.md §内联决策记录]（决策记录）、[31-acceptance-criteria.md]（验收标准）。

## 架构文档的核心原则

### 记录问题而非仅记录解决方案

架构源于对问题的深刻理解。记录问题的目标、约束和质量属性。

**好的做法**:
```markdown
## 问题

系统必须支持：
- 预期用户负载
- 响应时间要求
- 可用性要求

当前约束：
- 团队规模：5 人
- 部署环境：AWS
- 预算：每月 $5000

> **Why**: 这些约束决定了架构选择。单体架构足以支持当前负载，
> 微服务的复杂性在当前团队规模下不值得。
```

**不好的做法**:
```markdown
## 架构

我们使用微服务架构。每个服务独立部署。服务之间通过 HTTP 通信。
```

> **Why**: 只记录解决方案，未来工程师无法理解为什么做出这个选择，也无法判断何时应该改变。

### 架构简洁说明架构质量高

好的架构可以在两分钟内解释清楚。无法简洁描述，说明架构过于复杂。

**示例**:
```markdown
## 架构概述

三层架构：
- **Presentation** — HTTP handlers，请求验证，响应序列化
- **Domain** — 业务逻辑，领域模型，业务规则
- **Infrastructure** — 数据库访问，外部 API 调用，文件系统

依赖方向：Presentation → Domain ← Infrastructure

> **Why**: 清晰的分层使业务逻辑独立于技术细节。Domain 层可以独立测试，
> 不依赖数据库或 HTTP 框架。
```

## C4 模型

使用 C4 模型的四个层次描述架构。每个层次回答不同的问题。

### Level 1: 系统上下文（Context）

**回答**: 系统与外部的交互边界是什么？

**示例**:
```markdown
## 系统上下文

### 用户
- **终端用户** — 通过 Web 浏览器访问
- **管理员** — 通过管理后台管理配置

### 外部系统
- **支付网关** — 处理支付交易（Stripe API）
- **邮件服务** — 发送通知邮件（SendGrid）
- **身份提供商** — 用户认证（Auth0）

### 数据流
用户 → [系统] → 支付网关
用户 ← [系统] ← 邮件服务
```

### Level 2: 容器（Container）

**回答**: 系统由哪些可部署单元组成？

**示例**:
```markdown
## 容器

### Web API
- **技术**: Go 1.25, net/http
- **职责**: 处理 HTTP 请求，业务逻辑
- **部署**: Docker 容器，AWS ECS

### PostgreSQL
- **版本**: 15
- **职责**: 持久化业务数据
- **部署**: AWS RDS

### Redis
- **版本**: 7
- **职责**: 会话存储，缓存
- **部署**: AWS ElastiCache

> **Why**: 单个 API 服务足以支持当前负载。Redis 缓存减少数据库查询，
> 提升响应速度。
```

### Level 3: 组件（Component）

**回答**: 容器内部如何组织？

**示例**:
```markdown
## 组件（Web API）

```
/cmd/api/          — 应用入口
/internal/
  /handlers/       — HTTP handlers（Presentation 层）
  /domain/         — 业务逻辑和领域模型（Domain 层）
  /repository/     — 数据访问（Infrastructure 层）
  /middleware/     — HTTP 中间件（认证、日志、错误处理）
```

依赖规则：
- handlers 依赖 domain
- repository 实现 domain 定义的接口
- domain 不依赖任何其他层

> **Why**: 六边形架构。Domain 层是核心，不依赖外部技术。
> 便于测试和替换基础设施。
```

### Level 4: 代码（Code）

**回答**: 关键类和接口如何设计？通常不在 SPEC 中详细记录，仅记录关键接口和设计模式。

**示例**:
```markdown
## 关键接口

### Repository 接口

```go
// UserRepository defines the contract for user data access.
// Implementations must be safe for concurrent use.
type UserRepository interface {
    // FindByID retrieves a user by ID. Returns an error if not found.
    FindByID(ctx context.Context, id string) (*User, error)
    // Save persists a user. Returns an error if the operation fails.
    Save(ctx context.Context, user *User) error
}
```

实现：
- `PostgresUserRepository` — 生产环境
- `InMemoryUserRepository` — 测试环境

> **Why**: 接口在 domain 层定义，实现在 infrastructure 层。
> Domain 层不依赖具体数据库技术。
```

## 架构约束

明确记录必须遵守的架构规则，使用工具强制执行。

### 依赖方向约束

**示例**:
```markdown
## 依赖规则

### 禁止的依赖
- `domain` 包不得导入 `handlers` 或 `repository`
- `handlers` 包不得导入 `repository`（通过 domain 接口访问）

### 允许的依赖
- `handlers` → `domain`
- `repository` → `domain`（实现接口）
- `middleware` → `domain`（访问领域类型）

> **Why**: 保持 domain 层的独立性。业务逻辑不应依赖 HTTP 框架或数据库技术。

验证方法：
```bash
# Check domain package imports for forbidden dependencies
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domain/... | grep -E '(handlers|repository)'
# Should produce no output if dependencies are correct
```
```

### 命名约束

**示例**:
```markdown
## 命名约束

### 禁止的后缀
- `Manager` — 职责不清晰
- `Helper` — 杂物抽屉
- `Util` — 缺乏内聚性

### 推荐的模式
- `Repository` — 数据访问
- `Service` — 业务逻辑协调（仅在 domain 层）
- `Handler` — HTTP 请求处理（仅在 handlers 层）

> **Why**: 清晰的命名反映清晰的职责。模糊的后缀暗示设计问题。
```

### 质量属性场景

用 BDD 场景验证架构质量属性：

```gherkin
Scenario: 认证和授权
  Given 用户未登录
  When 用户访问受保护的资源
  Then 系统返回 401 Unauthorized
  And 响应不包含敏感信息

Scenario: 数据库切换
  Given 系统使用 PostgreSQL
  When 需要切换到 MySQL
  Then 只需实现新的 Repository
  And Domain 层代码无需修改
```

> **Why**: 质量属性场景验证架构决策的有效性（如六边形架构的可替换性）。

## 架构决策记录（ADR）

架构决策内联到 SPEC 中，使用 `> **Why**` 和 `> **Rejected**` 格式。

### 简单决策

**示例**:
```markdown
## 数据库选择

使用 PostgreSQL 15。

> **Why**: 需要 JSONB 支持和事务保证。团队已有 PostgreSQL 经验。
>
> **Rejected**:
> - MySQL — JSONB 支持较弱
> - MongoDB — 缺乏事务保证
> - SQLite — 不支持并发写入
```

### 复杂决策

使用决策表格式：

**示例**:
```markdown
## 架构风格决策

| 方面 | 选择 | 原因 | 拒绝的替代方案 |
|------|------|------|--------------|
| 架构风格 | 单体应用 | 团队 5 人，负载 3000 用户，单体足够 | 微服务（过度复杂，团队太小） |
| 分层模式 | 六边形架构 | Domain 层独立，易于测试 | 传统三层（Domain 依赖基础设施） |
| 通信方式 | HTTP/REST | 简单，工具支持好 | gRPC（增加复杂性，收益不明显） |
| 数据库 | PostgreSQL | JSONB + 事务 + 团队经验 | MongoDB（缺乏事务），MySQL（JSONB 弱） |

> **Why**: 当前阶段优先考虑简单性和团队效率。单体应用足以支持未来 2-3 年增长。
> 当团队扩大到 15+ 人或负载增长 10 倍时，重新评估架构。
```

## 架构视图

按需补充以下视图：

- **逻辑视图** — 功能分解和职责划分（核心领域 vs 支撑领域）
- **部署视图** — 部署拓扑和运行环境
- **数据视图** — 数据流和存储分布

> **Why**: 核心领域是业务价值所在，投入最多资源。支撑领域可以外包或使用现成方案。

## Terminology

| Term | Definition |
|------|-----------|
| **C4 模型** | 四层架构描述法：Context → Container → Component → Code |
| **ADR** | 架构决策记录（Architecture Decision Record），内联到 SPEC 中 |
| **质量属性** | 可量化的系统特性（响应时间、可用性、吞吐量） |
| **架构约束** | 必须遵守的架构规则，可通过工具强制执行 |

## Forbidden

- **Don't 只画图不写约束**: 架构图展示结构但不定义规则 → 图表 + 依赖规则 + 验证方法
- **Don't 记录实现细节**: 具体类实现和算法细节属于代码 → 只记录关键接口、设计模式、架构约束
- **Don't 包含部署步骤**: 部署操作属于运维文档 → 记录部署架构（拓扑、组件），不记录操作步骤
- **Don't 记录所有设计决策**: 局部决策在代码注释中记录 → 只记录架构级别的决策（影响多个组件）
- **Don't 使用模糊的质量属性**: "系统应该快"无法验证 → 使用具体场景（"98% 请求在 1 秒内响应"）
- **Don't 记录显而易见的约束**: "代码应该可读"不是架构约束 → 只记录非显而易见的、容易被违反的约束

## References

- [11-structure.md] — SPEC 结构：决策内联格式
- [20-content-boundaries.md] — 内容边界：什么属于 SPEC
- [11-structure.md §内联决策记录] — 决策记录方法
- [31-acceptance-criteria.md] — BDD 场景编写
