# SPEC 模板库

## Overview

提供不同类型 SPEC 的可复用模板。每个模板包含标准结构、必填章节、填写指导。

不涵盖：模板使用流程（见 README）、工具生成方法、自动化配置。

相关 SPEC：[11-structure.md]（SPEC 结构）、[21-writing-style.md]（写作风格）、[51-spec-examples.md]（完整示例）。

## 架构 SPEC 模板

### 适用场景

定义系统架构、记录架构决策、说明组件关系和依赖规则。

### 模板

```markdown
# <系统名称> 架构

## Overview

[2-3 句话]：系统范围和职责、边界、相关 SPEC。

## Architecture Views

### Context View

用户：
- **[角色名称]** — [职责和使用场景]

外部系统：
- **[系统名称]** — [交互目的]（[协议]）

数据流：
```
┌─────────┐
│  User   │
└────┬────┘
     │ HTTPS
┌────▼────┐      ┌──────────┐
│ System  │─────▶│ External │
└─────────┘      └──────────┘
```

> **Why**: [为什么这样划分边界]
> **Rejected**: [替代方案] — [原因]

### Container View

**[容器名称]**
- 技术：[技术栈和版本]
- 职责：[核心职责]
- 通信：[与其他容器的通信方式]

## Architecture Constraints

| 约束 | 说明 | 理由 |
|------|------|------|
| [约束名称] | [具体规则] | [为什么需要] |

## Quality Attributes

## Terminology

- **[术语]** — [定义]

## Forbidden

- **Don't [禁止的做法]**: [原因] → [正确做法]

## References

- [相关 SPEC]
```

### 填写指导

1. **Context View**：画出系统边界，明确外部依赖
2. **Container View**：列出所有可独立部署的单元
3. **Architecture Constraints**：写可验证的规则，避免模糊约束
4. **Quality Attributes**：Given-When-Then 格式，包含具体数字
5. 每个非显而易见的选择用 `> **Why**` 记录

## API SPEC 模板

### 适用场景

定义 API 端点契约：请求/响应格式、错误码、验收标准。

### 模板

```markdown
# <API 名称>

## Overview

[2-3 句话]：API 范围、边界、相关 SPEC。

## Endpoint: [HTTP 方法] [路径]

### Request

#### 路径参数
- `[参数名]` ([类型], required/optional) — [说明]

#### 查询参数
- `[参数名]` ([类型], optional, default: [默认值]) — [说明]

#### 请求体

**TypeScript**:
```typescript
interface [RequestName] {
  [field]: [type];  // [约束说明]
}
```

**Go**:
```go
type [RequestName] struct {
    [Field] [type] `json:"[json_name]"`
}
```

### Response

**TypeScript**:
```typescript
interface [ResponseName] {
  [field]: [type];
}
```

**Go**:
```go
type [ResponseName] struct {
    [Field] [type] `json:"[json_name]"`
}
```

### Errors

| Code | Condition | Response |
|------|-----------|----------|
| [状态码] | [触发条件] | `{"error": "[type]", "message": "[消息]"}` |

> **Why**: [为什么这样设计]
> **Rejected**: [替代方案] — [原因]

## Acceptance Criteria

Scenario: [场景名称]
  Given [前置条件]
  When [API 调用]
  Then [预期结果]

## Terminology

- **[术语]** — [定义]

## Forbidden

- **Don't [禁止的做法]**: [原因] → [正确做法]

## References

- [相关 SPEC]
```

### 填写指导

1. **类型对齐**：TypeScript 和 Go 字段名、类型、约束保持一致
2. **错误响应**：表格列出所有错误码，包含触发条件
3. **验收标准**：覆盖正常路径、边界条件、异常情况

## 数据模型 SPEC 模板

### 适用场景

定义核心业务实体、前后端数据结构对齐、验证规则和不变量。

### 模板

```markdown
# <实体名称> 数据模型

## Overview

[2-3 句话]：实体用途、边界、相关 SPEC。

## Schema

### 字段定义

| 字段 | 类型 | 必填 | 约束 | 说明 |
|------|------|------|------|------|
| [field] | [type] | 是/否 | [约束] | [说明] |

### TypeScript

```typescript
interface [EntityName] {
  [field]: [type];
}
```

### Go

```go
type [EntityName] struct {
    [Field] [type] `json:"[json_name]" validate:"[rules]"`
}
```

> **Why**: [为什么这样设计字段]

## Validation Rules

### 字段级验证

| 字段 | 规则 | 错误消息 |
|------|------|----------|
| [field] | [规则] | [消息] |

### 实体级验证

- [业务规则描述]

## Invariants

1. **[不变量名称]** — [具体规则]

## Relationships

### [关系名称]
- **类型**：[一对一/一对多/多对多]
- **外键**：[字段] → [目标实体].[字段]
- **级联**：[级联规则]

## State Machine（如适用）

| 状态 | 说明 | 可转换到 |
|------|------|---------|
| [state] | [说明] | [目标状态] |

## Terminology

- **[术语]** — [定义]

## Forbidden

- **Don't [禁止的做法]**: [原因] → [正确做法]

## References

- [相关 SPEC]
```

### 填写指导

1. **类型对齐**：TypeScript 和 Go 的字段名、类型、约束必须一致
2. **验证规则**：表格列出所有验证，包含错误消息
3. **不变量**：写出必须始终为真的业务规则

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **模板** | 可复用的 SPEC 结构，包含标准章节和填写指导 | Not "示例"（示例是完整文档） |
| **必填章节** | 确保 SPEC 完整性的章节，不可省略 | Not "可选章节" |

## Forbidden

- **Don't modify required sections**: 必填章节确保 SPEC 完整性 → 如果不适用，说明原因
- **Don't pad with filler content**: 模板是指导不是强制 → 只写有价值的内容
- **Don't copy-paste without modification**: 占位符必须替换 → 根据实际情况填写
- **Don't mix multiple templates**: 一个 SPEC 一个主题 → 多个主题创建多个 SPEC
- **Don't skip decision records**: `> **Why**` 是模板的核心 → 非显而易见的选择必须解释

## References

- [11-structure.md §内联决策记录] — 决策记录方法
- [20-content-boundaries.md] — 内容边界
- [21-writing-style.md] — 写作风格
- [31-acceptance-criteria.md] — 验收标准编写
- [40-architecture-specs.md] — 架构 SPEC 详细规范
- [41-api-specs.md] — API SPEC 详细规范
- [42-data-model-specs.md] — 数据模型 SPEC 详细规范
- [51-spec-examples.md] — 完整示例
