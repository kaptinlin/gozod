# 编码规范 SPEC

## Overview

定义如何为技术项目编写编码规范（Coding Standards）。编码规范指导 Agent 生成符合团队风格、最佳实践和语言惯例的代码。

适用范围：所有编码项目（Go、TypeScript、Python、Rust 等）。非编码项目不需要。

不涵盖：项目架构设计（见 [40-architecture-specs.md]）、API 契约（见 [41-api-specs.md]）、部署配置。

相关 SPEC：[11-structure.md]（SPEC 结构）、[20-content-boundaries.md]（内容边界）、[21-writing-style.md]（写作风格）。

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Must Follow** | 必须遵守的规则，违反导致审查不通过 | 不是建议 |
| **Forbidden** | 明确禁止的模式，必须提供替代方案 | 不是"不推荐" |
| **信任层级** | 根据数据来源决定验证策略 | 不是所有数据都需要验证 |

## 必需章节

### 1. 核心理念（Core Principles）

建立团队共识，定义"好代码"的标准：设计哲学、信任模型、权衡取舍。

**示例**（TypeScript）：
```markdown
### 三个信任层级

| 层级 | 数据来源 | 验证方式 |
|------|----------|----------|
| **完全信任** | 内部函数、自有 API | TypeScript 类型足够 |
| **部分信任** | 外部 API、WebSocket | 类型 + 防御性编程 `?.` `??` |
| **零信任** | JWT/Token、CLI 输入 | **必须** Zod 运行时验证 |
```

**示例**（Go）：
```markdown
- **KISS**: 三行相似代码优于单次使用的抽象
- **DRY**: 优先使用标准库（slices, maps, cmp, errors）
- **YAGNI**: 基于证据优化（先 profile 再加 sync.Pool）
```

### 2. 规则分类（Rule Categories）

按优先级和领域组织：CRITICAL > HIGH > MEDIUM > LOW。

**示例**（Go）：
```markdown
| Priority | Category | Guide |
|----------|----------|-------|
| 1 | Naming | SPECS/44-naming-rules.md |
| 2 | Error Handling | SPECS/45-error-handling.md |
| 3 | Design Patterns | SPECS/46-design-patterns.md |
| 4 | Concurrency | SPECS/47-concurrency.md |
```

### 3. 具体规则（Specific Rules）

每条规则包含：场景、约束（✅/❌）、原因、代码示例。

**示例**（TypeScript）：
```markdown
### 运行时验证

**场景**：处理外部数据（JWT、CLI 输入、外部 API）

**约束**：
- ✅ 使用 Zod 验证 JWT payload
- ❌ 直接信任 `jwt.decode()` 返回值

**示例**：
\`\`\`typescript
const TokenPayload = z.object({
  sub: z.string().uuid(),
  exp: z.number().int().positive()
});
const payload = TokenPayload.parse(jwt.decode(token));
\`\`\`
```

### 4. 禁止项（Forbidden）

使用统一格式列出：

**示例**（Go）：
```markdown
- **Don't use `panic` for normal errors** — Use `error` returns; panic only for invariants
- **Don't copy `sync.Mutex`** — Never copy types with pointer methods
```

**示例**（TypeScript 表格式）：
```markdown
| 禁止项 | 原因 | 替代方案 |
|--------|------|----------|
| `any` 类型 | 破坏类型安全 | `unknown` + 类型守卫 |
| `@ts-ignore` | 隐藏问题 | 修复类型错误 |
```

### 5. 检查清单（Checklist）

代码审查前的自检列表：

```markdown
#### 类型安全
- [ ] 没有使用 `any` 类型
- [ ] 没有使用 `@ts-ignore`

#### 运行时验证（按需）
- [ ] 🔴 JWT/Token 使用 Zod 验证（必须）
- [ ] 🟠 CLI 输入使用 Zod 验证（建议）
- [ ] ⚪ 自有 API 信任 TypeScript（不需要）
```

### 可选章节

根据语言特性按需添加：性能优化、并发模式、测试规范、工具配置。

## 编写原则

### 现代化优先

拥抱最新稳定特性，淘汰过时模式。

**版本要求示例**：Go 1.26+、TypeScript 5.7+、React 19.2+、Node.js 22+ LTS

**淘汰清单示例**：
- Go：`interface{}` → `any`，`io/ioutil` → `io`/`os`
- TypeScript：`namespace` → ES modules，`enum` → union types
- React：`forwardRef` → ref prop，class components → function components

### 可执行性

规则具体到 Agent 能直接应用。

```markdown
❌ 模糊：错误响应应该包含错误信息。

✅ 可执行：所有 API 错误返回 RFC 7807 Problem Details：
\`\`\`json
{"type": "https://api.example.com/errors/not-found", "title": "Resource Not Found", "status": 404}
\`\`\`
```

### 最小化

只包含必须遵守的规则。

```markdown
❌ 所有函数必须有注释
✅ 公开 API 必须有 JSDoc/godoc

❌ 所有错误必须包装
✅ 跨包边界的错误必须包装（Go）
```

### 分层信任

根据数据来源决定验证策略，避免过度验证。

| 层级 | 数据来源 | 验证策略 |
|------|----------|----------|
| **完全信任** | 内部函数、自有 API | 类型系统足够 |
| **部分信任** | 外部 API、WebSocket | 类型 + 防御性编程 |
| **零信任** | JWT/Token、CLI 输入 | 运行时验证 |

### 决策透明

每个非显而易见的规则解释原因：

```markdown
### 错误包装

跨包边界的错误必须包装：`return fmt.Errorf("failed to save user: %w", err)`

> **Why**: 包装错误提供调用栈上下文，便于调试
> **Rejected**: 使用 pkg/errors — 标准库 errors 已足够（Go 1.13+）
```

## 文件组织

**单文件**（< 500 行）：所有规则在 `SPECS/43-coding-standards.md`

**多文件**（> 500 行）：按领域拆分，43 为索引：
```
SPECS/
├── 43-coding-standards.md  (索引和核心理念)
├── 44-naming.md
├── 45-error-handling.md
├── 46-design-patterns.md
└── 47-concurrency.md
```

## Forbidden

- **Don't include architecture design**: 架构属于 [40-architecture-specs.md] → 只定义代码风格和模式
- **Don't include API contracts**: API 契约属于 [41-api-specs.md] → 只定义如何实现 API
- **Don't write vague rules**: 规则必须可执行 → 提供具体代码示例和约束
- **Don't over-specify**: 只包含必须遵守的规则 → 信任标准库和框架默认行为
- **Don't write rules for hypothetical needs**: 基于真实问题 → 按需添加
- **Don't skip decision records**: 非显而易见的规则用 `> **Why**` 和 `> **Rejected**` 记录

## References

- [11-structure.md] — SPEC 结构规范
- [20-content-boundaries.md] — 内容边界
- [21-writing-style.md] — 写作风格
- [24-code-examples.md] — 代码示例规范
- [40-architecture-specs.md] — 架构 SPEC
- [41-api-specs.md] — API SPEC
