package gozod_test

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

func ExampleToJSONSchema() {
	document, err := gozod.ToJSONSchema(gozod.String())
	if err != nil {
		panic(err)
	}
	fmt.Println(document.Type[0])
	// Output: string
}

func ExampleToJSONSchemaRegistry() {
	registry := gozod.NewRegistry[gozod.GlobalMeta]().
		Add(gozod.String(), gozod.GlobalMeta{ID: "Name"})
	document, err := gozod.ToJSONSchemaRegistry(registry)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(document.Defs))
	// Output: 1
}
