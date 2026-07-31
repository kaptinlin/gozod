# Error Customization

GoZod resolves validation messages when an issue is finalized. Choose the
narrowest owner that knows the right message: a schema or check for
domain-specific text, a parse context for one request, and global configuration
only for process-wide defaults such as locale.

## Inspect Validation Errors

```go
_, err := gozod.String().Parse(42)

var zodErr *gozod.ZodError
if gozod.IsZodError(err, &zodErr) {
    for _, issue := range zodErr.Issues {
        fmt.Printf("%s: %s\n", issue.Code, issue.Message)
    }
}
```

Each finalized `gozod.ZodIssue` includes its code, message, path, and
check-specific fields such as `Minimum`, `Maximum`, `Expected`, and `Received`.

## Schema and Check Messages

Constructors and fluent checks accept a message or an error map:

```go
schema := gozod.String("Name must be text").
    Min(2, "Name must contain at least two characters")

dynamic := gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
    return fmt.Sprintf("Password is too short; minimum is %v", issue.Properties["minimum"])
})
```

Use schema-level messages when the rule itself owns the wording. They have the
highest precedence.

## Per-Parse Messages

`Parse` accepts an optional `*core.ParseContext`. Build one explicitly and pass
it to the same parse method:

```go
ctx := core.NewParseContext().WithCustomError(func(issue core.ZodRawIssue) string {
    return fmt.Sprintf("request value failed %s", issue.Code)
})

result, err := schema.Parse(input, ctx)
```

Parse contexts are immutable by convention: `WithCustomError` and
`WithReportInput` return a new context. A per-parse message is used only when a
schema or check did not already provide one.

## Global Defaults

Set process-wide defaults with `gozod.SetConfig`:

```go
gozod.SetConfig(&gozod.ZodConfig{
    CustomError: func(issue gozod.ZodRawIssue) string {
        return fmt.Sprintf("validation failed: %s", issue.Code)
    },
})
```

Global configuration affects later parses across the process. Set it during
application initialization rather than per request. Pass `nil` to reset the
configuration to its zero state.

For localized defaults:

```go
gozod.SetConfig(locales.ZhCN())
```

## Message Precedence

GoZod uses the first non-empty message in this order:

1. Schema or check message.
2. Per-parse `core.ParseContext` error map.
3. Global `ZodConfig.CustomError`.
4. Global `ZodConfig.LocaleError`.
5. Built-in default message.

An error map may return an empty string to defer to the next level.

```go
schema := gozod.String("schema message")
ctx := core.NewParseContext().WithCustomError(func(core.ZodRawIssue) string {
    return "request message"
})

_, err := schema.Parse(42, ctx)
// err uses "schema message".
```

Use [Error Formatting](error-formatting.md) to turn the resulting
`*gozod.ZodError` into human-readable, tree, or flat output.
