package gozod_test

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
)

func ExamplePrettifyError() {
	schema := gozod.Object(gozod.ObjectSchema{
		"name": gozod.String().Min(2),
	})
	_, err := schema.Parse(map[string]any{"name": "A"})
	zodErr, ok := errors.AsType[*gozod.ZodError](err)
	if !ok {
		panic(err)
	}

	pretty := gozod.PrettifyError(zodErr)
	tree := gozod.TreeifyError(zodErr)
	flat := gozod.FlattenError(zodErr)

	fmt.Println(len(pretty) > 0)
	fmt.Println(len(tree.Properties["name"].Errors))
	fmt.Println(len(flat.FieldErrors["name"]))
	// Output:
	// true
	// 1
	// 1
}

func ExampleString_parseContext() {
	ctx := core.NewParseContext().WithCustomError(func(core.ZodRawIssue) string {
		return "request-specific message"
	})

	_, err := gozod.String().Parse(42, ctx)
	zodErr, ok := errors.AsType[*gozod.ZodError](err)
	if !ok {
		panic(err)
	}
	fmt.Println(zodErr.Issues[0].Message)
	// Output: request-specific message
}
