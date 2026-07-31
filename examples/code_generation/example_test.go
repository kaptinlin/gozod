package main

import "fmt"

func ExampleUser_Schema() {
	schema := User{}.Schema()
	user, err := schema.Parse(User{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "Alice",
		Email:  "alice@example.com",
		Age:    28,
		Status: "active",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(user.Name)
	// Output: Alice
}
