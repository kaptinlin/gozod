# TypeScript Zod v4 参考代码

这个目录包含作为 git submodule 的 TypeScript Zod v4 源代码，用作 GoZod 开发的准确参考。

## 目录结构

```
.reference/
└── zod/                                    # Zod 主仓库 (git submodule)
    └── packages/zod/src/v4/               # Zod v4 源代码
        ├── core/                          # 核心实现
        │   ├── api.ts                     # 主要 API 定义
        │   ├── checks.ts                  # 验证检查实现
        │   ├── schemas.ts                 # Schema 定义
        │   ├── errors.ts                  # 错误处理
        │   ├── util.ts                    # 工具函数
        │   └── ...                        # 其他核心文件
        ├── locales/                       # 国际化文件
        │   ├── en.ts                      # 英语
        │   ├── zh-CN.ts                   # 简体中文
        │   └── ...                        # 其他语言
        └── index.ts                       # v4 入口文件
```

## 使用指南

### 1. 初始化 Submodule

如果你是第一次克隆 GoZod 仓库：

```bash
git submodule update --init --recursive
```

### 2. 更新 Submodule

获取 Zod 的最新更新：

```bash
git submodule update --remote .reference/zod
```

### 3. 查找参考代码

在编写 GoZod 代码时，请参考相应的 TypeScript 文件：

- **检查系统**: `.reference/zod/packages/zod/src/v4/core/checks.ts`
- **Schema 类型**: `.reference/zod/packages/zod/src/v4/core/schemas.ts`
- **错误处理**: `.reference/zod/packages/zod/src/v4/core/errors.ts`
- **API 设计**: `.reference/zod/packages/zod/src/v4/core/api.ts`

## 重要说明

⚠️ **只参考 v4 代码**：GoZod 基于 Zod v4，请不要参考 v3 目录下的代码。

✅ **精确对应**：在 GoZod 代码注释中引用的 TypeScript 代码应该与这些文件中的实际代码完全匹配。

📚 **文档一致性**：按照 `.cursor/rules/typescript-to-go-comments.mdc` 中的规范格式引用 TypeScript 代码。

## 代码对应示例

当你在 GoZod 中实现某个功能时，应该这样引用：

```go
// ZodCheckDef defines the configuration for validation checks
// TypeScript original code:
//
//	export interface $ZodCheckDef {
//	  check: string;
//	  error?: errors.$ZodErrorMap<never> | undefined;
//	  abort?: boolean | undefined;
//	}
type ZodCheckDef struct {
	Check string       // Check type identifier
	Error *ZodErrorMap // Custom error mapping  
	Abort bool         // Whether to abort on validation failure
}
```

其中 TypeScript 代码应该从 `.reference/zod/packages/zod/src/v4/core/checks.ts` 中精确复制。

## 链接

- 原始仓库: https://github.com/colinhacks/zod
- Zod v4 源码: https://github.com/colinhacks/zod/tree/main/packages/zod/src/v4 
