# Error Formatting

GoZod returns validation failures as `*gozod.ZodError`. Pass that error object
directly to the formatting helpers; the helpers preserve issue messages and
paths.

## Extract a Validation Error

```go
value, err := schema.Parse(input)
if err != nil {
    var zodErr *gozod.ZodError
    if !gozod.IsZodError(err, &zodErr) {
        return err
    }
    // Format zodErr here.
}
```

`ZodError.Issues` remains available for direct inspection, but it is not the
argument to the formatting helpers.

## PrettifyError

Use `PrettifyError` for logs, command-line output, and other human-readable
messages:

```go
fmt.Println(gozod.PrettifyError(zodErr))
```

## TreeifyError

Use `TreeifyError` when a consumer needs errors arranged like the validated
object or collection:

```go
tree := gozod.TreeifyError(zodErr)
if name := tree.Properties["name"]; name != nil {
    fmt.Println(name.Errors)
}
```

Object fields are stored in `Properties`; collection indexes are stored in
`Items`. Each node can also contain errors that apply to the node itself.

## FlattenError

Use `FlattenError` for flat forms and API responses that group errors by the
first path segment:

```go
flat := gozod.FlattenError(zodErr)
fmt.Println(flat.FormErrors)
fmt.Println(flat.FieldErrors["email"])
```

`FormErrors` contains issues with an empty path. `FieldErrors` groups all other
issues by their first path segment, so use `TreeifyError` when nested path
detail matters.

## Custom Mapping

Tree and flat output accept a mapper over finalized `gozod.ZodIssue` values:

```go
mapper := func(issue gozod.ZodIssue) string {
    return fmt.Sprintf("%s: %s", issue.Code, issue.Message)
}

tree := gozod.TreeifyErrorWithMapper(zodErr, mapper)
flat := gozod.FlattenErrorWithMapper(zodErr, mapper)
```

For most applications, set messages on the schema or parse context before
formatting. See [Error Customization](error-customization.md).

## Path Utilities

```go
fmt.Println(gozod.ToDotPath([]any{"users", 0, "email"}))
// users[0].email
```

`FormatErrorPath` also supports `"dot"` and `"bracket"` styles when an external
protocol requires a specific representation.
