# TypeScript Zod v4 Reference Code

This directory contains the TypeScript Zod v4 source code as a git submodule, serving as an accurate reference for GoZod development.

## Directory Structure

```
.reference/
└── zod/                                    # Zod main repository (git submodule)
    └── packages/zod/src/v4/               # Zod v4 source code
        ├── core/                          # Core implementation
        │   ├── api.ts                     # Main API definitions
        │   ├── checks.ts                  # Validation checks implementation
        │   ├── schemas.ts                 # Schema definitions
        │   ├── errors.ts                  # Error handling
        │   ├── util.ts                    # Utility functions
        │   └── ...                        # Other core files
        ├── locales/                       # Internationalization files
        │   ├── en.ts                      # English
        │   ├── zh-CN.ts                   # Simplified Chinese
        │   └── ...                        # Other languages
        └── index.ts                       # v4 entry file
```

## Usage Guide

### 1. Initialize Submodule

If you are cloning the GoZod repository for the first time:

```bash
git submodule update --init --recursive
```

### 2. Update Submodule

To get the latest updates from Zod:

```bash
git submodule update --remote .reference/zod
```

### 3. Find Reference Code

When writing GoZod code, please refer to the corresponding TypeScript files:

- **Check System**: `.reference/zod/packages/zod/src/v4/core/checks.ts`
- **Schema Types**: `.reference/zod/packages/zod/src/v4/core/schemas.ts`
- **Error Handling**: `.reference/zod/packages/zod/src/v4/core/errors.ts`
- **API Design**: `.reference/zod/packages/zod/src/v4/core/api.ts`

## Important Notes

⚠️ **Reference v4 Code Only**: GoZod is based on Zod v4. Please do not refer to the code in the v3 directory.

✅ **Exact Correspondence**: The TypeScript code referenced in GoZod code comments should exactly match the actual code in these files.

📚 **Documentation Consistency**: Follow the format specified in `.cursor/rules/typescript-to-go-comments.mdc` for referencing TypeScript code.

## Code Correspondence Example

When implementing a feature in GoZod, you should reference it like this:

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

The TypeScript code here should be copied exactly from `.reference/zod/packages/zod/src/v4/core/checks.ts`.

## Links

- Original repository: https://github.com/colinhacks/zod
- Zod v4 source code: https://github.com/colinhacks/zod/tree/main/packages/zod/src/v4 
