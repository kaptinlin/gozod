---
description: Analyze reference projects and generate root ANALYSIS.md defining research scope and priorities. Use when planning research direction, analyzing reference repositories, or structuring research topics.
name: research-analyzing
---


# Research Analyzing

分析 `.references/` 目录中的参考项目，生成根目录 `ANALYSIS.md` 调研方向规划，定义调研主题（Topics）、参考项目（References）、关键文件（Key Files）和调研重点（Focus Areas）。

**核心原则**：只分析方向，不深入细节。ANALYSIS.md 是调研的"地图"，不是调研报告本身。

**使用时机**：在调研阶段开始时，需要规划调研方向和识别参考项目。

## 工作流位置

```
.references/          →  research-analyzing  →  ANALYSIS.md (可选审核)
(参考项目 submodules)                                ↓
                                            research-tasking  →  TODO.yaml
                                                                      ↓
                                                                 执行任务
                                                                      ↓
                                                                .research/*.md
                                                                (调研报告)
```

## 何时使用

使用此技能当：
- 用户说"分析调研方向"、"规划调研范围"、"制定调研计划"
- 开始新的调研项目，需要识别参考项目和关键文件
- `.references/` 目录已包含参考项目（git submodules）
- 需要生成调研方向规划供团队审核

不使用此技能当：
- 已有 `ANALYSIS.md`，需要生成任务清单（使用 `research-tasking`）
- 需要执行具体调研任务（使用 `research-writing`）
- 需要深入分析源码（那是 `research-writing` 的职责）

## 输入

**必需**：
- 项目根目录路径
- `.references/` 目录（包含参考项目 git submodules）

**可选**：
- 现有的调研文档或需求说明
- 特定的调研主题或领域

## 输出

在项目根目录生成 `ANALYSIS.md`，包含：

1. **调研主题**：需要调研的方向和领域
2. **参考项目**：每个主题对应的参考项目路径
3. **关键文件**：每个参考项目中需要重点查看的文件路径
4. **调研重点**：每个主题需要关注的问题（不是答案）

## 执行步骤

### Step 1: 扫描参考项目

遍历 `.references/` 目录，识别所有参考项目：

```bash
# 列出所有 submodules
cd /path/to/project
git submodule status

# 或直接查看目录结构
ls -la .references/
```

对每个参考项目，记录：
- 项目名称和路径
- 项目类型（从目录结构或 README 推断）
- 主要功能领域

**注意**：此步骤只识别项目类型和功能领域，不分析具体实现方式。

### Step 2: 识别调研主题

基于参考项目的分布和类型，识别调研主题：

**覆盖性要求**：
- 必须覆盖 `.references/` 下的所有参考项目
- 每个参考项目至少归入一个调研主题
- 如果某个项目难以归类，单独创建主题

**主题提取原则**：
- 按功能领域分组（如 Chart 库、Table 库、Editor 库）
- 按技术维度分组（如跨平台共享、WC 桥接）
- 主题粒度适中（不要过细或过粗）

**注意**：此步骤只提取主题名称和方向，不在 ANALYSIS.md 中描述具体实现方式。

### Step 3: 识别关键文件

对每个参考项目，列出需要重点查看的文件路径：

**识别方法**：
- 查看项目 README 中提到的核心文件
- 查看 `package.json` 的 `main`/`exports` 字段
- 查看目录结构中的 `src/core/`、`src/hooks/`、`packages/core/` 等
- 浏览关键文件了解其职责和作用
- 列出典型的入口文件和核心模块

**注意**：需要读取文件了解其作用，但在 ANALYSIS.md 中只列出文件路径和简要说明，不描述具体实现方式。

### Step 4: 定义调研重点

对每个主题，列出需要关注的问题（不是答案）：

**问题类型**：
- 架构分层：如何分离 State/Behavior/UI？
- 跨平台策略：哪些代码可以共享？
- API 设计：核心接口是什么样的？
- 性能优化：使用了哪些优化手段？
- 设计模式：观察到哪些架构模式？

**注意**：需要浏览代码了解项目的技术方向，但在 ANALYSIS.md 中只列出调研问题，不回答问题，不描述具体实现。答案由 `research-writing` 提供。

### Step 5: 生成 ANALYSIS.md

在项目根目录创建 `ANALYSIS.md`：

```markdown
# Research Analysis

## 调研主题

### 1. {主题名称}

**调研方向**：
- {方向1}
- {方向2}

**参考项目**：
- `{项目路径1}` — {项目简介}
- `{项目路径2}` — {项目简介}

**关键文件**：
- `{项目路径1}/{文件路径}` — {文件说明}
- `{项目路径2}/{文件路径}` — {文件说明}

**调研重点**：
- {问题1}
- {问题2}

---

### 2. {主题名称}

...

## 覆盖性检查

确认所有 `.references/` 下的项目均已覆盖：

| 参考项目 | 归入主题 |
|----------|----------|
| `{项目路径1}` | 主题1 |
| `{项目路径2}` | 主题1, 主题2 |
| ... | ... |

**未覆盖项目**: 无（如有遗漏，必须说明原因）
```

## 完整示例

**场景**：规划 Chart/Table/Editor 库的调研方向

**输入**：
```
.references/
├── domain-specific/
│   ├── visx/
│   ├── victory/
│   ├── recharts/
│   ├── ag-grid/
│   ├── prosekit/
│   └── lexical/
├── headless-behavior/
│   ├── tanstack-table/
│   └── tanstack-virtual/
└── design-system/
    ├── gluestack-ui/
    └── tamagui/
```

**输出 ANALYSIS.md**：
```markdown
# Research Analysis

## 调研主题

### 1. Chart 库调研

**调研方向**：
- 现代图表库的架构分层模式（State/Behavior/UI）
- Canvas/SVG/WebGL 渲染策略及跨平台抽象
- 数据处理层（scale、axis、layout）的共享边界
- 交互层（tooltip、zoom、pan）的平台适配策略

**参考项目**：
- `.references/domain-specific/visx/` — Airbnb React 图表库
- `.references/domain-specific/victory/` — Formidable 跨平台图表库
- `.references/domain-specific/recharts/` — 流行 React 图表库

**关键文件**：
- `visx/packages/visx-scale/src/scales/` — Scale 实现
- `victory/packages/victory-core/src/victory-util/` — 核心工具
- `recharts/src/chart/` — 图表组件

**调研重点**：
- 如何分离数据处理和渲染逻辑？
- SVG 渲染如何跨 Web/RN 平台？
- 动画系统如何统一接口？

---

### 2. Table 库调研

**调研方向**：
- TanStack Table 的 headless 架构设计
- State 层、Behavior 层的分离模式
- 跨框架适配策略（React/Vue/Solid）
- Plugin 架构（14 个 feature 模块的组合）

**参考项目**：
- `.references/headless-behavior/tanstack-table/` — TanStack Table 核心
- `.references/headless-behavior/tanstack-virtual/` — 虚拟滚动
- `.references/domain-specific/ag-grid/` — 企业级表格

**关键文件**：
- `tanstack-table/packages/table-core/src/core/table.ts` — Table 工厂
- `tanstack-table/packages/table-core/src/features/` — Feature 模块
- `tanstack-table/packages/react-table/src/index.tsx` — React 适配

**调研重点**：
- Feature Plugin 如何组合？
- Row Model Pipeline 如何实现 lazy evaluation？
- 如何在 WC/React/RN 三端复用 table-core？

---

### 3. 跨平台共享策略调研

**调研方向**：
- Headless 库的跨平台模式（State/Behavior/UI 分层）
- State 层纯逻辑实现（100% 共享）
- Token 系统跨平台输出（DTCG JSON → CSS + TS）

**参考项目**：
- `.references/design-system/gluestack-ui/` — 跨 Web/RN 设计系统
- `.references/design-system/tamagui/` — 跨平台设计系统

**关键文件**：
- `gluestack-ui/packages/gluestack-core/src/button/aria/useButton.ts` — 纯 TS ARIA
- `gluestack-ui/packages/gluestack-utils/src/` — 共享工具库

**调研重点**：
- Gluestack UI 如何实现 70% 共享率？
- Token 系统如何双输出（CSS + TS）？
- 纯 TS 描述符如何作为跨平台载体？

## 覆盖性检查

| 参考项目 | 归入主题 |
|----------|----------|
| `.references/domain-specific/visx/` | 主题1 |
| `.references/domain-specific/victory/` | 主题1 |
| `.references/domain-specific/recharts/` | 主题1 |
| `.references/headless-behavior/tanstack-table/` | 主题2 |
| `.references/headless-behavior/tanstack-virtual/` | 主题2 |
| `.references/domain-specific/ag-grid/` | 主题2 |
| `.references/domain-specific/prosekit/` | （待添加 Editor 主题） |
| `.references/domain-specific/lexical/` | （待添加 Editor 主题） |
| `.references/design-system/gluestack-ui/` | 主题3 |
| `.references/design-system/tamagui/` | 主题3 |

**未覆盖项目**: prosekit、lexical（需添加 Editor 主题）
```

## Research Standards

ANALYSIS.md 应包含以下调研标准，供 research-writing 执行时参照：

### 语言要求

所有调研报告使用中文撰写（除非用户另有指定）。专有名词（项目名、API 名、技术术语）保留英文原文。

### 深度控制

**ANALYSIS.md 只规划方向，不写具体实现**：
- 可以读取代码了解项目的技术方向和架构思路
- 列出调研主题、参考项目、关键文件路径
- 列出调研重点问题（不回答问题）
- 不在 ANALYSIS.md 中描述具体实现方式、API 签名、代码模式
- 深入的实现分析由 `research-writing` 完成

### 报告模板

所有调研报告必须遵循 `references/report-template.md` 中定义的结构。

## 关键原则

1. **只规划方向，不写实现**：可以读代码了解技术方向，但 ANALYSIS.md 中不描述具体实现方式
2. **列出路径和作用**：记录关键文件路径和简要说明，但不分析实现细节
3. **提出问题，不回答问题**：列出调研重点问题，答案由 research-writing 提供
4. **结构清晰**：按主题组织，每个主题包含方向、项目、文件、重点
5. **覆盖完整**：确保所有参考项目都归入某个主题

## 下一步

生成 `ANALYSIS.md` 后：

1. **人工审核**（可选）：团队确认调研范围和优先级
2. **生成任务**：使用 `research-tasking` 技能基于 `ANALYSIS.md` 生成 `TODO.yaml`
3. **执行调研**：按 `TODO.yaml` 执行调研任务，生成 `.research/*.md` 报告

## 常见问题

**Q: ANALYSIS.md 和调研报告有什么区别？**

A:
- `ANALYSIS.md`（根目录）：调研方向规划，列出主题、项目、文件、问题
- `.research/*.md`：调研报告，回答 ANALYSIS.md 中提出的问题

**Q: 需要读取参考项目的源码吗？**

A: 需要浏览代码了解技术方向，但：
- 可以查看目录结构、README、关键文件
- 了解项目的架构思路和技术选型
- 但在 ANALYSIS.md 中不写具体实现方式
- 只列出文件路径、简要说明、调研问题

**Q: 如何确定关键文件？**

A:
- 查看 `package.json` 的 `main`/`exports` 字段
- 查看 README 中提到的核心模块
- 查看典型的目录结构（`src/core/`、`packages/core/` 等）

**Q: 调研重点应该写什么？**

A: 写问题，不写答案或实现：
- Good: "如何分离 State 和 UI 层？"
- Good: "SVG 渲染如何跨 Web/RN 平台？"
- Bad: "使用 hooks 分离 State 和 UI 层"（这是实现方式）
- Bad: "通过 react-native-svg 实现跨平台"（这是具体方案）
