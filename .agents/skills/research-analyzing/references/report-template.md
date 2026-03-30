# Research Report Template — Structure Definition

定义调研报告的标准结构，供 research-analyzing 规划调研范围和结构时参照。

## File Naming

- 文件名使用 kebab-case
- 按 ANALYSIS.md / TODO.yaml 中的 task ID 命名
- 示例：`.research/R01a-wc-framework-core-apis.md`

## Report Structure

```markdown
---
id: R__x
title: <报告标题>
task: <TODO.yaml 中的 task ID>
date: YYYY-MM-DD
status: draft              # draft | review | complete
scope:
  - <调研子问题 1>
  - <调研子问题 2>
tags: [<关键词>]
---

# R__x — <报告标题>

## 执行摘要                    ← 结论表格 + 置信度
## 1-N. <主题章节>             ← 按主题横切多个项目，含 API 提炼
## N-3. 对本项目的落地建议      ← 文字 + 建议接口设计
## N-2. 决策矩阵               ← 推荐/备选/否决 + 依据
## N-1. 代码块索引              ← 快速定位所有 API/接口
## N. 引用清单                 ← 按项目分组
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

- 格式：`<!-- ref: <项目短名>/<相对路径>:<行号> -->`
- 项目短名省略 `.references/` 前缀和分类目录
- 多源引用逗号分隔：`<!-- ref: material-web/...:27-42, spectrum-wc/...:34-40 -->`

## Code Block Rules

| 允许 | 不允许 |
|------|--------|
| 类型签名 (`type`, `interface`) | 具体算法实现 |
| 函数签名（参数 + 返回值） | 函数体 / 业务逻辑 |
| 从参考项目提炼的 API 摘要 | 直接复制源码 |
| Schema / Config 结构定义 | 运行时代码 / 样板代码 |
| CSS 变量契约 / 选择器模式 | 完整 CSS 文件 |
| 使用示例（调用侧 3-5 行） | 完整组件实现 |
