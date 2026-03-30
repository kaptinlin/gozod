# 数据模型 SPEC 规范

## Overview

定义如何编写数据模型相关的 SPEC：实体字段、类型约束、关系规则、验证规则、不变量。

不涵盖：存储实现（数据库 schema）、ORM 配置、数据迁移脚本。

相关 SPEC：[11-structure.md]（SPEC 结构）、[41-api-specs.md]（API 契约）、[12-naming.md]（命名规范）。

## 实体定义

每个实体必须包含：实体名称和用途、字段列表（类型+约束+必填标记）、不变量。

**示例**:
```markdown
## Token 实体

设计 Token 信息。

### 字段

| 字段 | 类型 | 必填 | 约束 | 说明 |
|------|------|------|------|------|
| id | UUID | 是 | 主键，唯一 | Token 唯一标识符 |
| name | string | 是 | 唯一，Token 名称格式 | Token 名称 |
| value | string | 是 | 非空 | Token 值 |
| category | enum | 是 | color/spacing/typography | Token 类别 |
| status | enum | 是 | active/deprecated/archived | Token 状态 |
| created_at | timestamp | 是 | 不可变 | 创建时间 |
| updated_at | timestamp | 是 | 自动更新 | 最后更新时间 |

> **Why**: UUID 作为主键避免枚举攻击。Name 唯一约束确保一个名称只能使用一次。

### 不变量
- Name 一旦设置不可为空
- `created_at` 创建后不可修改
- `updated_at` 在任何字段修改时自动更新
- 归档 Token 时 `status` 设为 `archived`，不物理删除

> **Why**: 软删除保留审计记录。`created_at` 不可变确保时间线准确性。
```

## 枚举与状态机

枚举类型必须定义所有值、说明、转换规则。

**示例**:
```markdown
### TokenStatus

| 值 | 说明 | 可转换到 |
|----|------|---------|
| active | 正常使用 | deprecated, archived |
| deprecated | 已废弃 | active, archived |
| archived | 已归档 | （终态） |

```
active ──┬──> deprecated ──> archived
         └──────────────────> archived
```

禁止的转换：archived → active, archived → deprecated

> **Why**: 明确的状态机防止非法状态转换。归档是终态，防止意外恢复。
>
> **Rejected**:
> - 添加 `inactive` 状态 — 与 `deprecated` 语义重叠
> - 允许从 `archived` 恢复 — 违反审计要求
```

## 关系定义

### 一对多

```markdown
## Token - Version 关系

### 关系类型
一对多（1:N）

### 外键
- `Version.token_id` → `Token.id`
- 级联删除：是（Token 删除时，所有版本也删除）

### 查询方向
- Token → Version：`token.Versions()`
- Version → Token：`version.Token()`

> **Why**: 版本是 Token 的历史记录，Token 删除时版本也应删除。
>
> **Rejected**: 不级联删除 — 保留孤立的版本没有意义
```

### 多对多

```markdown
## Token - Collection 关系

### 关联表 `token_collections`

| 字段 | 类型 | 约束 |
|------|------|------|
| token_id | UUID | 外键 → Token.id |
| collection_id | UUID | 外键 → Collection.id |
| added_at | timestamp | 添加时间 |
| added_by | UUID | 外键 → User.id |

- (token_id, collection_id) 组合唯一
- 删除 Token：删除关联记录
- 删除 Collection：阻止删除（如果有 Token 关联）

> **Why**: 关联表记录添加时间和操作人，用于审计。
```

### 自引用

```markdown
## Category 自引用

分类树结构（父子关系）。

- `parent_id` (UUID, nullable) — NULL 表示根分类
- 不允许循环引用
- 最大深度：5 层

> **Why**: 限制深度防止过深的层次结构影响性能和用户体验。
>
> **Rejected**:
> - 无限深度 — 可能导致性能问题和 UI 渲染困难
```

## 验证规则

### 字段级验证

```markdown
### Name
- 格式：`^[a-z]+\.[a-z]+\.[a-z]+$`
- 长度：最大 100 字符
- 唯一性：全局唯一（不区分大小写）

### Value
- 格式根据 category 验证
  - color: 十六进制颜色码或 RGB/RGBA
  - spacing: CSS 长度单位
  - typography: CSS 字体值

> **Why**: 严格的验证规则防止无效数据进入系统。
```

### 实体级验证

```markdown
## Component 实体验证

- 组件必须至少引用一个 Token
- 状态为 `published` 时，`published_at` 不能为空
- 状态为 `deprecated` 时，`deprecated_at` 必须晚于 `published_at`

### 验证时机
- 创建时、状态转换时、修改 Token 引用时

> **Why**: 实体级验证确保业务不变量。状态和时间戳的一致性确保业务流程正确。
```

### 跨实体验证

```markdown
### Token 引用检查
- 引用的 Token 必须存在且状态为 `active`
- 验证时机：组件创建前

### 权限检查
- 只有 Token 所有者或管理员可以修改
- 已发布的 Token 不能修改 category
```

## 不变量（Invariants）

不变量是必须始终为真的业务规则，违反意味着数据损坏。

```markdown
## Order 不变量

1. **总额一致性**
   - `order.total = SUM(order_items.quantity * order_items.price)`

2. **状态一致性**
   - `status = paid` → `paid_at IS NOT NULL`
   - `status = shipped` → `shipped_at > paid_at`

3. **库存一致性**
   - 创建订单时：`order_item.quantity <= product.stock`
   - 取消订单时：恢复库存

### 验证方式
- 应用层验证（创建/更新时）
- 数据库约束（CHECK 约束、触发器）
- 定期一致性检查（后台任务）
```

### 聚合根

```markdown
## 聚合根：Order

Order 是聚合根，负责维护内部一致性。

### 边界
- Order（根）→ OrderItem（子实体）→ ShippingAddress（值对象）

### 规则
- 外部只能通过 Order 修改 OrderItem
- Order 负责验证所有不变量

> **Why**: 聚合根模式确保不变量在单一位置维护，防止绕过验证。
>
> **Rejected**:
> - 允许直接操作子实体 — 无法保证不变量
> - 每个实体都是聚合根 — 过度设计
```

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **实体** | 有独立 ID 和生命周期的业务对象 | Not "值对象"（值对象无独立 ID） |
| **不变量** | 必须始终为真的业务规则 | Not "约束"（约束是字段级限制） |
| **聚合根** | 维护内部一致性的顶层实体 | Not "普通实体"（子实体不可直接操作） |

## Forbidden

- **Don't include ORM mapping**: ORM 配置是实现细节 → 只定义数据模式
- **Don't include query examples**: 查询逻辑属于 API SPEC → 定义关系和约束
- **Don't use DB-specific types**: 使用通用类型 → `string` 而非 `VARCHAR`
- **Don't expose implementation details**: 如自增 ID、表名 → 使用业务语言
- **Don't define all possible fields**: 只定义核心字段 → 次要字段在实现时添加

## References

- [11-structure.md §内联决策记录] — 决策记录
- [12-naming.md] — 命名规范
- [20-content-boundaries.md] — 内容边界
- [40-architecture-specs.md] — 架构约束
- [41-api-specs.md] — API 契约
