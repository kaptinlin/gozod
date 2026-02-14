# GoZod Go 1.26 Refactoring Guide

本文档分析 Go 1.26 两个关键新特性 `new(expr)` 和**泛型自我引用约束 (Self-Referential Generic Constraints)** 在 GoZod 代码库中的适用场景、具体改进位置和重构策略。

---

## 一、`new(expr)` 改进

### 1.1 特性回顾

Go 1.26 的 `new(expr)` 允许直接对表达式取地址创建指针，无需临时变量：

```go
// Before (Go 1.25)
v := someExpr
ptr := &v

// After (Go 1.26)
ptr := new(someExpr)
```

### 1.2 GoZod 已有的 `new(expr)` 使用

GoZod 已在多处采用了 `new(expr)` 模式（在之前的 Go 1.26 重构中完成）：

| 文件 | 示例 | 数量 |
|------|------|------|
| `types/iso.go:11-17` | `PrecisionMinute = new(-1)` 等精度常量 | 7 |
| `types/complex.go` | `new(complex64(v))`, `new(v)` | 12 |
| `types/float.go:584-586` | `new(float32(v))`, `new(v)` | 2 |
| `locales/*.go` | `return new(info)` (28个语言文件) | 28 |
| `cmd/gozodgen/writer.go:558` | `return new(rule)` | 1 |

### 1.3 仍可改进的临时变量取地址模式

以下位置仍存在 `tmp := expr; field = &tmp` 的模式，可用 `new(expr)` 简化：

#### 1.3.1 `types/bigint.go` — Refine 中的指针构造

```go
// 当前 (bigint.go:351-353)
if val, ok := v.(*big.Int); ok {
    ptr := &val
    return fn(any(ptr).(T))
}

// Go 1.26 改进
if val, ok := v.(*big.Int); ok {
    return fn(any(new(val)).(T))
}
```

同一模式在 `bigint.go:478-480`：

```go
// 当前
if val, ok := v.(*big.Int); ok {
    ptr := &val
    return any(ptr).(T), true
}

// Go 1.26 改进
if val, ok := v.(*big.Int); ok {
    return any(new(val)).(T), true
}
```

#### 1.3.2 `types/bool.go` — Check 方法中的值复制取地址

```go
// 当前 (bool.go:320-321)
bCopy := b
fn(any(&bCopy).(T), payload)

// Go 1.26 改进
fn(any(new(b)).(T), payload)
```

同一模式在 `bool.go:279-280` (Refine 方法中)：

```go
// 当前
cp := b
return fn(any(&cp).(T))

// Go 1.26 改进
return fn(any(new(b)).(T))
```

#### 1.3.3 `types/stringbool.go` — 类型 switch 取地址

```go
// 当前 (stringbool.go:546-548)
case StringBoolOptions:
    opts = &v
    rest = params[1:]

// Go 1.26 改进
case StringBoolOptions:
    opts = new(v)
    rest = params[1:]
```

#### 1.3.4 `types/text.go` — JWTOptions 取地址

```go
// 当前 (text.go:156-157)
if o, ok := p.(JWTOptions); ok {
    opts = &o
}

// Go 1.26 改进
if o, ok := p.(JWTOptions); ok {
    opts = new(o)
}
```

#### 1.3.5 `types/enum.go` — Refine/Check 中的值复制

```go
// 当前 (enum.go:294-295)
if val, ok := v.(T); ok {
    return fn(any(&val).(R))
}

// Go 1.26 改进
if val, ok := v.(T); ok {
    return fn(any(new(val)).(R))
}
```

同一模式出现在 `enum.go:342-343` (Check 方法中)。

#### 1.3.6 `types/bool.go` — Prefault/PrefaultFunc 中的取地址

```go
// 当前 (bool.go:177-178)
case *bool:
    in.SetPrefaultValue(&v)

// Go 1.26 改进
case *bool:
    in.SetPrefaultValue(new(v))
```

同一模式在 `bool.go:193` (PrefaultFunc)。

### 1.4 `new(expr)` 改进汇总

| 文件 | 行号 | 当前模式 | 改进 |
|------|------|----------|------|
| `types/bigint.go` | 351-353 | `ptr := &val` | `new(val)` |
| `types/bigint.go` | 478-480 | `ptr := &val` | `new(val)` |
| `types/bool.go` | 279-280 | `cp := b; &cp` | `new(b)` |
| `types/bool.go` | 320-321 | `bCopy := b; &bCopy` | `new(b)` |
| `types/bool.go` | 177 | `&v` in Prefault | `new(v)` |
| `types/bool.go` | 193 | `&v` in PrefaultFunc | `new(v)` |
| `types/enum.go` | 294-295 | `&val` | `new(val)` |
| `types/enum.go` | 342-343 | `cp := v; &cp` | `new(v)` |
| `types/stringbool.go` | 547 | `opts = &v` | `opts = new(v)` |
| `types/text.go` | 157 | `opts = &o` | `opts = new(o)` |

**注意**: 许多类似的模式（如各 schema 类型的 Refine/Check/Prefault 方法中的临时变量取地址）遍布整个 `types/` 目录。上表列出了代表性案例，完整重构应 grep 搜索所有 `&val`、`&cp`、`&v` 等模式。

---

## 二、泛型自我引用约束

### 2.1 特性回顾

Go 1.26 允许定义自我引用的泛型约束：

```go
type Builder[Self Builder[Self]] interface {
    WithName(string) Self
    Build() any
}
```

这使得接口方法的返回类型可以是实现者自身的具体类型，而不是接口类型。

### 2.2 GoZod 中的核心痛点：34 种 Schema 类型的链式 API 重复

GoZod 拥有 **34 种 schema 类型**，每种都独立实现了一套近乎相同的链式方法：

| Schema 类型 | 方法数 | 文件 |
|-------------|--------|------|
| `ZodObject` | 63 | `types/object.go` |
| `ZodString` | 55 | `types/string.go` |
| `ZodStruct` | 49 | `types/struct.go` |
| `ZodFloatTyped` | 48 | `types/float.go` |
| `ZodIntegerTyped` | 46 | `types/integer.go` |
| `ZodComplex` | 44 | `types/complex.go` |
| `ZodRecord` | 44 | `types/record.go` |
| `ZodMap` | 42 | `types/map.go` |
| `ZodLazy` | 42 | `types/lazy.go` |
| `ZodArray` | 42 | `types/array.go` |
| `ZodSet` | 41 | `types/set.go` |
| `ZodNetwork` | 40 | `types/network.go` |
| `ZodIds` | 40 | `types/ids.go` |
| `ZodBigInt` | 40 | `types/bigint.go` |
| `ZodSlice` | 39 | `types/slice.go` |
| `ZodFile` | 39 | `types/file.go` |
| `ZodStringBool` | 37 | `types/stringbool.go` |
| `ZodFunction` | 36 | `types/function.go` |
| `ZodEnum` | 36 | `types/enum.go` |
| `ZodTime` | 34 | `types/time.go` |
| `ZodBool` | 34 | `types/bool.go` |
| `ZodText` | — | `types/text.go` |
| `ZodEmail` | — | `types/email.go` |
| `ZodIso` | — | `types/iso.go` |
| `ZodLiteral` | — | `types/literal.go` |
| `ZodNever` | — | `types/never.go` |
| `ZodNil` | — | `types/nil.go` |
| `ZodAny` | — | `types/any.go` |
| `ZodUnknown` | — | `types/unknown.go` |
| `ZodUnion` | — | `types/union.go` |
| `ZodXor` | — | `types/xor.go` |
| `ZodIntersection` | — | `types/intersection.go` |
| `ZodDiscriminatedUnion` | — | `types/discriminated_union.go` |
| `ZodTuple` | — | `types/tuple.go` |

**合计: ~1400+ 个方法，其中大量是结构相同但类型签名不同的重复代码。**

### 2.3 三大重复模式

#### 模式 A：`withCheck` / `withInternals` / `withPtrInternals`

所有 34 种类型都有这三个内部 helper，逻辑完全相同，仅类型参数不同：

```go
// ZodBool (types/bool.go:353-373)
func (z *ZodBool[T]) withCheck(check core.ZodCheck) *ZodBool[T] {
    in := z.internals.Clone()
    in.AddCheck(check)
    return z.withInternals(in)
}

func (z *ZodBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodBool[*bool] {
    return &ZodBool[*bool]{internals: &ZodBoolInternals{
        ZodTypeInternals: *in,
        Def:              z.internals.Def,
    }}
}

func (z *ZodBool[T]) withInternals(in *core.ZodTypeInternals) *ZodBool[T] {
    return &ZodBool[T]{internals: &ZodBoolInternals{
        ZodTypeInternals: *in,
        Def:              z.internals.Def,
    }}
}

// ZodEnum (types/enum.go:437-461) — 完全相同的模式
func (z *ZodEnum[T, R]) withCheck(check core.ZodCheck) *ZodEnum[T, R] {
    in := z.internals.Clone()
    in.AddCheck(check)
    return z.withInternals(in)
}

func (z *ZodEnum[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodEnum[T, *T] {
    return &ZodEnum[T, *T]{internals: &ZodEnumInternals[T]{
        ZodTypeInternals: *in,
        Def:              z.internals.Def,
        Entries:          z.internals.Entries,
        Values:           z.internals.Values,
    }}
}

func (z *ZodEnum[T, R]) withInternals(in *core.ZodTypeInternals) *ZodEnum[T, R] {
    return &ZodEnum[T, R]{internals: &ZodEnumInternals[T]{
        ZodTypeInternals: *in,
        Def:              z.internals.Def,
        Entries:          z.internals.Entries,
        Values:           z.internals.Values,
    }}
}
```

#### 模式 B：共同修饰符方法

每种类型都重复实现的修饰符方法（仅返回类型不同）：

```go
// 以下方法在 34 种类型中完全重复（仅类型名不同）
func (z *Zod<Type>[T]) Optional() *Zod<Type>[*<base>]
func (z *Zod<Type>[T]) Nilable() *Zod<Type>[*<base>]
func (z *Zod<Type>[T]) Nullish() *Zod<Type>[*<base>]
func (z *Zod<Type>[T]) ExactOptional() *Zod<Type>[T]
func (z *Zod<Type>[T]) Default(v <base>) *Zod<Type>[T]
func (z *Zod<Type>[T]) DefaultFunc(fn func() <base>) *Zod<Type>[T]
func (z *Zod<Type>[T]) Prefault(v <base>) *Zod<Type>[T]
func (z *Zod<Type>[T]) PrefaultFunc(fn func() <base>) *Zod<Type>[T]
func (z *Zod<Type>[T]) NonOptional() *Zod<Type>[<base>]
func (z *Zod<Type>[T]) Meta(meta core.GlobalMeta) *Zod<Type>[T]
func (z *Zod<Type>[T]) Describe(desc string) *Zod<Type>[T]
```

每个修饰符方法的内部逻辑都是：

```
1. in := z.internals.Clone()
2. in.Set<Something>(value)
3. return z.withInternals(in)  // 或 z.withPtrInternals(in)
```

#### 模式 C：通用验证/组合方法

```go
func (z *Zod<Type>[T]) Refine(fn func(T) bool, params ...any) *Zod<Type>[T]
func (z *Zod<Type>[T]) RefineAny(fn func(any) bool, params ...any) *Zod<Type>[T]
func (z *Zod<Type>[T]) Check(fn func(T, *core.ParsePayload), params ...any) *Zod<Type>[T]
func (z *Zod<Type>[T]) With(fn func(T, *core.ParsePayload), params ...any) *Zod<Type>[T]
func (z *Zod<Type>[T]) And(other any) *ZodIntersection[any, any]
func (z *Zod<Type>[T]) Or(other any) *ZodUnion[any, any]
func (z *Zod<Type>[T]) Transform(fn func(<base>, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any]
func (z *Zod<Type>[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any]
```

### 2.4 Go 1.26 自我引用约束设计

#### 2.4.1 核心 Schema 接口约束

```go
// core/constraints.go (新文件)

// SchemaChain 定义所有 schema 类型共享的链式修饰符合约。
// Self 参数通过自我引用约束确保返回类型的正确性。
type SchemaChain[Self SchemaChain[Self]] interface {
    // 元数据方法
    Describe(string) Self
    Meta(GlobalMeta) Self

    // 验证修饰符 — 返回自身类型
    ExactOptional() Self

    // 内部方法（可不导出）
    withCheck(ZodCheck) Self
    withInternals(*ZodTypeInternals) Self
}
```

#### 2.4.2 Copy-on-Write 接口约束

```go
// CopyOnWrite 定义支持 Copy-on-Write 语义的 schema 约束。
type CopyOnWrite[Self CopyOnWrite[Self]] interface {
    // Clone 返回一个新的 schema 实例，内部状态独立
    withCheck(ZodCheck) Self
    withInternals(*ZodTypeInternals) Self
}
```

#### 2.4.3 实际价值评估

自我引用约束在 GoZod 中的价值主要是**编译期一致性保证**，而非代码复用（因为 Go 的泛型不支持通过接口提供默认方法实现）。

**可以做到的：**
- 编译期保证所有 schema 类型实现了统一的链式 API
- 为泛型函数提供精确的约束，避免运行时类型断言
- 定义操作 schema 的通用函数（如 batch validation、schema composition）

**无法做到的（Go 限制）：**
- 无法通过接口提供默认方法实现来消除代码重复
- 无法用 embedding 来共享泛型方法（因为返回类型是具体类型）
- `Optional()` 和 `Nilable()` 改变了类型参数（从 `T` 到 `*base`），无法用简单的自我引用约束表达

### 2.5 自我引用约束的具体应用场景

#### 场景 1：泛型 Schema 操作函数

```go
// 当前：无法编写类型安全的通用 schema 操作
func ApplyDefaults(schema any, defaults map[string]any) any {
    // 需要类型断言，不安全
}

// Go 1.26：类型安全的泛型操作
type Describable[S Describable[S]] interface {
    Describe(string) S
    Meta(core.GlobalMeta) S
    Internals() *core.ZodTypeInternals
}

func WithDescription[S Describable[S]](schema S, desc string) S {
    return schema.Describe(desc)
}

func WithMeta[S Describable[S]](schema S, meta core.GlobalMeta) S {
    return schema.Meta(meta)
}
```

#### 场景 2：Schema 验证管道组合

```go
// 类型安全的验证管道组合
type Refineable[S Refineable[S]] interface {
    RefineAny(func(any) bool, ...any) S
    Internals() *core.ZodTypeInternals
}

// 批量应用自定义验证规则
func ApplyRefinements[S Refineable[S]](schema S, rules []func(any) bool) S {
    for _, rule := range rules {
        schema = schema.RefineAny(rule)
    }
    return schema
}
```

#### 场景 3：Registry 链式 API 约束

```go
// core/registry.go 已有链式方法，可以用约束来泛化
type ChainableRegistry[M any, Self ChainableRegistry[M, Self]] interface {
    Add(ZodSchema, M) Self
    Remove(ZodSchema) Self
}

// 使用示例：确保 Registry 实现了链式 API
var _ ChainableRegistry[GlobalMeta, *Registry[GlobalMeta]] = (*Registry[GlobalMeta])(nil)
```

#### 场景 4：Schema 工厂的类型安全约束

```go
// 确保所有 schema 工厂函数返回的类型都实现了完整的 API
type FullSchema[S FullSchema[S]] interface {
    // 核心
    Internals() *core.ZodTypeInternals
    IsOptional() bool
    IsNilable() bool
    ParseAny(any, ...*core.ParseContext) (any, error)

    // 修饰符
    ExactOptional() S
    Describe(string) S
    Meta(core.GlobalMeta) S
}

// 编译期验证
var _ FullSchema[*ZodBool[bool]] = (*ZodBool[bool])(nil)
var _ FullSchema[*ZodString[string]] = (*ZodString[string])(nil)
// ... 对所有 34 种类型
```

### 2.6 与 TypeScript Zod v4 的对比

TypeScript Zod v4 使用 **wrapper 类型**来实现修饰符 (`$ZodOptional`, `$ZodNullable`, `$ZodDefault`)，每个修饰符是一个独立的 schema 类型。

GoZod 的设计**故意避免了 wrapper 类型**，而是通过 Copy-on-Write 模式直接在同一类型上修改泛型参数（如 `ZodBool[bool]` → `ZodBool[*bool]`）。这意味着：

- TypeScript 的继承/mixin 模式无法直接移植到 Go
- Go 的自我引用约束提供了 TypeScript 没有的**编译期接口一致性保证**
- GoZod 的 Copy-on-Write 模式比 Zod v4 的 wrapper 模式更高效（少一层间接）

---

## 三、重构优先级与影响评估

### 3.1 优先级矩阵

| 重构项 | 优先级 | 影响范围 | 风险 | 改进类型 | 状态 |
|--------|--------|----------|------|----------|------|
| `new(expr)` 临时变量清理 | **P1** | 27 个文件 | 低 | 代码简洁性 | ✅ 已完成 |
| 编译期 Schema 接口约束 | **P2** | core/ + types/ | 低 | 类型安全 | ✅ 已完成 |
| 泛型 Schema 操作函数 | **P3** | internal/, pkg/ | 中 | 代码复用 | 待定 |

### 3.2 P1: `new(expr)` 清理（立即可执行）

**目标**: 消除所有残余的 `tmp := expr; &tmp` 模式。

**操作步骤**:

1. 全局搜索模式：
   ```bash
   # 搜索临时变量取地址
   grep -rn "ptr := &\|cp := \|bCopy := \|opts = &" types/ --include="*.go" | grep -v test
   ```

2. 逐个替换为 `new(expr)` 形式。

3. 运行完整测试套件确认无破坏：
   ```bash
   make test && make lint
   ```

**影响文件**: `types/bigint.go`, `types/bool.go`, `types/enum.go`, `types/stringbool.go`, `types/text.go` 及其他类似模式的 schema 文件。

### 3.3 P2: 编译期 Schema 约束（架构改进）

**目标**: 在 `core/` 中定义自我引用约束接口，为所有 schema 类型添加编译期一致性检查。

**操作步骤**:

1. 创建 `core/constraints.go`:

```go
package core

// Describable 保证所有实现者都有 Describe 和 Meta 方法，
// 并且返回正确的具体类型。
type Describable[S Describable[S]] interface {
    Describe(string) S
    Meta(GlobalMeta) S
    Internals() *ZodTypeInternals
}
```

2. 在每种 schema 类型文件末尾添加编译期验证：

```go
// types/bool.go 末尾
var _ core.Describable[*ZodBool[bool]] = (*ZodBool[bool])(nil)

// types/string.go 末尾
var _ core.Describable[*ZodString[string]] = (*ZodString[string])(nil)
```

3. 逐步扩展约束范围。

### 3.4 P3: 泛型操作函数（可选改进）

**目标**: 利用自我引用约束编写通用的 schema 操作工具函数。

这一步依赖 P2 的约束定义，可在需要时实现。典型用例：

- `func WithDescription[S Describable[S]](schema S, desc string) S`
- `func ApplyChecks[S Checkable[S]](schema S, checks []core.ZodCheck) S`
- `func BatchValidate[S Parseable[S]](schemas []S, input any) []error`

---

## 四、不建议的重构方向

### 4.1 不应尝试用 embedding 消除方法重复

Go 的泛型嵌入（generic embedding）无法解决 GoZod 的方法重复问题，因为：

- 每个链式方法必须返回**具体类型**（如 `*ZodBool[T]`），而非接口类型
- Go 不支持接口默认方法实现
- 嵌入的方法无法访问外层类型的泛型参数

```go
// 这在 Go 中是不可能的
type BaseSchema[Self any, T any] struct { ... }

func (b *BaseSchema[Self, T]) Describe(desc string) Self {
    // 无法构造 Self 的实例
}
```

### 4.2 不应引入代码生成来消除重复

虽然代码生成可以减少手写量，但会引入维护成本，且 GoZod 的方法重复在逻辑上足够简单，不值得增加构建复杂度。

### 4.3 不应改变 Copy-on-Write 为 Wrapper 类型

TypeScript Zod v4 使用 `$ZodOptional` 等 wrapper 类型，但 GoZod 的 Copy-on-Write 模式是**有意的设计选择**，提供：

- 更少的类型间接层
- 更低的内存分配
- 更简单的类型推断
- 与 Go 惯用法更一致

---

## 五、两个特性在 GoZod 中的互补关系

| 特性 | GoZod 中的典型场景 | 影响范围 | 改进类型 |
|------|-------------------|----------|----------|
| `new(expr)` | 类型转换取地址、Refine/Check 中值复制取地址、构造器参数取地址 | ~10 处源文件 | 微观代码简洁性 |
| 泛型自我引用 | 链式 Builder 模式的接口统一、Copy-on-Write helper 约束、Schema 操作泛型函数 | 架构级 | 宏观类型安全 |

这两个特性解决的是不同层面的问题：

- **`new(expr)`** 是**表达式级别**的语法改进，消除临时变量，提升代码密度
- **泛型自我引用约束**是**类型系统级别**的表达力提升，为 34 种 schema 类型的 ~1400 个方法提供编译期一致性保证

两者在同一个项目中同时展示，从微观到宏观覆盖了 Go 1.26 对验证库/Schema Builder 类项目的全部改进价值。

---

## 六、参考对照

### 6.1 TypeScript Zod v4 类型层次

参见 `.reference/zod/packages/zod/src/v4/core/schemas.ts`:

- `$ZodType` — 基础 schema 接口（包含 `_zod` 内部状态）
- `$ZodOptional<T>` — Optional wrapper（GoZod 中用 Copy-on-Write 替代）
- `$ZodNullable<T>` — Nullable wrapper（GoZod 中用 Copy-on-Write 替代）
- `$ZodDefault<T>` — Default wrapper（GoZod 中用 Copy-on-Write 替代）

### 6.2 GoZod 架构层次

参见 `core/interfaces.go`:

- `ZodType[T]` — 泛型 schema 接口 (Parse, MustParse, Internals, IsOptional, IsNilable)
- `ZodSchema` — 非泛型 schema 接口 (ParseAny, Internals)
- `ZodTypeInternals` — 共享的内部状态结构体（含 Clone 方法用于 Copy-on-Write）

### 6.3 文件位置快速索引

| 关注点 | 文件位置 |
|--------|----------|
| Schema 接口定义 | `core/interfaces.go` |
| Copy-on-Write 实现 | `core/interfaces.go:106-122` (Clone) |
| Registry 链式 API | `core/registry.go:35-57` |
| Schema 实现模板 | `.agents/rules/schema_implementation_guide.mdc` |
| 编码规范 | `.agents/rules/coding-standards.mdc` |
| 项目结构 | `.agents/rules/project-structure.mdc` |
