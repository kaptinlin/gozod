# Research Report Template — Writing Guide

执行 research-writing 时参照本模板产出报告。

## Report Structure

```markdown
---
id: R__x
title: <报告标题>
task: <task ID>
date: YYYY-MM-DD
status: draft              # draft | review | complete
scope:
  - <调研子问题 1>
  - <调研子问题 2>
tags: [<关键词>]
---

# R__x — <报告标题>

## 执行摘要

| # | 结论 | 置信度 |
|---|------|--------|
| 1 | **<一句话结论>** | 高/中/低 |
| 2 | **<一句话结论>** | 高/中/低 |

---

## 1. <主题章节>

按调研子问题组织，每个章节围绕一个主题横切多个参考项目。

### 1.1 <子节>

正文分析。源码引用使用行内短标注：
<!-- ref: project-name/path/to/file.ts:10-25 -->

### 1.2 <子节>

> **API 提炼**（从参考项目源码中提炼的核心签名）：

​```typescript
// 提炼自 <项目名>（简化签名，非原始源码）
interface XxxApi {
  method(param: Type): ReturnType;
}
​```

---

## N-3. 对本项目的落地建议

每条建议包含文字说明 + 建议的 API / 接口设计。

### 建议 1：<标题>

<文字说明>

​```typescript
// 建议接口（非实现代码）
interface SuggestedApi {
  // ...类型签名、函数签名、Schema 结构
}
​```

### 建议 2：<标题>

<文字说明>

​```css
/* CSS 契约示例（如涉及样式策略） */
:host {
  --component-token: var(--semantic-token, var(--primitive-token, <fallback>));
}
​```

## N-2. 决策矩阵

| 决策点 | 推荐 | 备选 | 否决 | 依据 |
|--------|------|------|------|------|
| <问题> | <方案> | <方案> | <方案 + 理由> | §X |

## N-1. 代码块索引

| 位置 | 类型 | 描述 |
|------|------|------|
| §1.2 | 参考提炼 | `XxxApi` — <项目名> 的核心签名 |
| §N-3 建议 1 | 建议接口 | `SuggestedApi` — 本项目适配器 |

## N. 引用清单

### <项目名>
| 路径 | 行号 | 用途 |
|------|------|------|
| `<相对路径>` | XX-YY | <引用目的> |
```

## Heading Hierarchy

- `#` — 报告标题（仅 1 个）
- `##` — 一级章节（执行摘要、正文主题、决策矩阵、引用清单）
- `###` — 二级子节
- `####` — 三级子节（尽量少用）

**禁止**：`## 2.1` 跟在 `## 2` 后面（应为 `### 2.1`）

## Report Organization

按**主题维度**横切多个项目，而不是按项目逐个罗列：

- Good: `## 响应式 Property 系统` → 在一个章节内对比 Lit / Stencil / Vanilla / FAST
- Bad: `## Lit 分析` → 列出所有 Lit 特性；`## Stencil 分析` → 列出所有 Stencil 特性

## Source Reference Format

使用 HTML 注释行内短标注，紧跟在论据句后：

```markdown
Material 把 token 落到 CSS 自定义属性，按 ref/sys/component 分层。
<!-- ref: material-web/docs/theming/README.md:27-42 -->
```

**规范**：
- 格式：`<!-- ref: <项目短名>/<相对路径>:<行号> -->`
- 项目短名省略 `.references/` 前缀和分类目录
- 多源引用逗号分隔：`<!-- ref: material-web/...:27-42, spectrum-wc/...:34-40 -->`

## Code Block Rules

报告不包含实现代码，但必须包含 API / 接口设计：

| 允许 | 不允许 |
|------|--------|
| 类型签名 (`type`, `interface`) | 具体算法实现 |
| 函数签名（参数 + 返回值） | 函数体 / 业务逻辑 |
| 从参考项目提炼的 API 摘要 | 直接复制源码 |
| Schema / Config 结构定义 | 运行时代码 / 样板代码 |
| CSS 变量契约 / 选择器模式 | 完整 CSS 文件 |
| 使用示例（调用侧 3-5 行） | 完整组件实现 |

### API Extraction（参考提炼）

从参考项目源码提炼最小核心签名，帮读者快速理解 API 形态：

```typescript
// 提炼自 Zag connect 模式（简化签名摘要）
type ConnectFn<TContext, TApi> = (
  service: Service<TContext>,
  normalizeProps: NormalizePropsFn
) => TApi;

interface SelectApi {
  getTriggerProps(): ElementAttrs;
  getContentProps(): ElementAttrs;
  getItemProps(item: Item): ElementAttrs;
}
```

### Suggested API（建议接口）

基于分析结论，在"落地建议"章节中给出可执行的接口草案：

```typescript
// 建议接口：WC 适配器
interface ZagController<TContext, TApi> {
  hostConnected(): void;
  hostDisconnected(): void;
  readonly api: TApi;
  send(event: EventObject): void;
}
```
