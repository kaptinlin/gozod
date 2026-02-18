package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/gozod"
)

// Example of circular reference handling with GoZod struct tags
// This demonstrates how GoZod automatically detects and handles circular references
// using lazy evaluation to avoid stack overflow errors.

// User represents a user with friends (circular reference)
type User struct {
	Name    string  `gozod:"required,min=2,max=50"`
	Email   string  `gozod:"required,email"`
	Age     int     `gozod:"min=18,max=120"`
	Friends []*User `gozod:"min=0"` // Circular reference - User references itself
}

// TreeNode represents a tree structure with parent and children
type TreeNode struct {
	Value    string      `gozod:"required"`
	Parent   *TreeNode   `gozod:""` // Circular reference to parent
	Children []*TreeNode `gozod:""` // Circular reference to children
}

// Department and Employee demonstrate mutual circular references
type Department struct {
	Name      string      `gozod:"required,min=2"`
	Manager   *Employee   `gozod:""`
	Employees []*Employee `gozod:"min=0"`
}

type Employee struct {
	Name       string      `gozod:"required,min=2"`
	Department *Department `gozod:""` // Circular reference back to Department
	Supervisor *Employee   `gozod:""` // Self-reference
	Team       []*Employee `gozod:""` // Self-reference array
}

func main() {
	fmt.Println("=== Circular Reference Handling Example ===")
	fmt.Println()

	// Example 1: Simple self-referencing structure
	demoSelfReference()
	fmt.Println()

	// Example 2: Tree structure with parent/child references
	demoTreeStructure()
	fmt.Println()

	// Example 3: Mutual references between types
	demoMutualReferences()
}

func demoSelfReference() {
	fmt.Println("1. Self-Referencing User Structure:")
	fmt.Println("------------------------------------")

	// Create schema from struct with circular references
	// GoZod automatically detects the circular reference and uses lazy evaluation
	schema := gozod.FromStruct[User]()

	// Create a user with friends
	user := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
		Friends: []*User{
			{
				Name:  "Bob",
				Email: "bob@example.com",
				Age:   30,
				Friends: []*User{
					{
						Name:    "Charlie",
						Email:   "charlie@example.com",
						Age:     28,
						Friends: nil, // Terminate the chain
					},
				},
			},
		},
	}

	// Validate the user
	result, err := schema.Parse(user)
	if err != nil {
		log.Fatal("Validation failed:", err)
	}

	fmt.Printf("✓ User validated successfully: %s\n", result.Name)
	fmt.Printf("  Friends: %d\n", len(result.Friends))
	if len(result.Friends) > 0 {
		fmt.Printf("  First friend: %s\n", result.Friends[0].Name)
		if len(result.Friends[0].Friends) > 0 {
			fmt.Printf("  Friend's friend: %s\n", result.Friends[0].Friends[0].Name)
		}
	}

	// Test validation error
	invalidUser := User{
		Name:  "A", // Too short, min=2
		Email: "invalid-email",
		Age:   17, // Too young, min=18
	}

	_, err = schema.Parse(invalidUser)
	if err != nil {
		fmt.Println("✓ Validation correctly failed for invalid user")
		if zodErr, ok := errors.AsType[*gozod.ZodError](err); ok {
			fmt.Printf("  Issues found: %d\n", len(zodErr.Issues))
		}
	}
}

func demoTreeStructure() {
	fmt.Println("2. Tree Structure with Parent/Child References:")
	fmt.Println("-----------------------------------------------")

	schema := gozod.FromStruct[TreeNode]()

	// Create a tree structure
	root := TreeNode{
		Value:  "Root",
		Parent: nil,
		Children: []*TreeNode{
			{
				Value:  "Child1",
				Parent: nil, // In real usage, would point back to root
				Children: []*TreeNode{
					{
						Value:    "Grandchild1",
						Parent:   nil,
						Children: nil,
					},
				},
			},
			{
				Value:    "Child2",
				Parent:   nil,
				Children: nil,
			},
		},
	}

	result, err := schema.Parse(root)
	if err != nil {
		log.Fatal("Tree validation failed:", err)
	}

	fmt.Printf("✓ Tree validated successfully: %s\n", result.Value)
	fmt.Printf("  Children: %d\n", len(result.Children))
	if len(result.Children) > 0 && len(result.Children[0].Children) > 0 {
		fmt.Printf("  Grandchild: %s\n", result.Children[0].Children[0].Value)
	}
}

func demoMutualReferences() {
	fmt.Println("3. Mutual References Between Types:")
	fmt.Println("-----------------------------------")

	// Department and Employee have mutual references
	deptSchema := gozod.FromStruct[Department]()
	empSchema := gozod.FromStruct[Employee]()

	// Create a department
	dept := Department{
		Name:    "Engineering",
		Manager: nil,
		Employees: []*Employee{
			{
				Name:       "Alice",
				Department: nil, // Would normally point back to dept
				Supervisor: nil,
				Team:       nil,
			},
			{
				Name:       "Bob",
				Department: nil,
				Supervisor: nil,
				Team:       []*Employee{},
			},
		},
	}

	deptResult, err := deptSchema.Parse(dept)
	if err != nil {
		log.Fatal("Department validation failed:", err)
	}

	fmt.Printf("✓ Department validated: %s\n", deptResult.Name)
	fmt.Printf("  Employees: %d\n", len(deptResult.Employees))

	// Create an employee hierarchy
	ceo := Employee{
		Name:       "CEO",
		Department: nil,
		Supervisor: nil,
		Team: []*Employee{
			{
				Name:       "VP Engineering",
				Department: nil,
				Supervisor: nil, // Would normally point back to CEO
				Team:       nil,
			},
			{
				Name:       "VP Sales",
				Department: nil,
				Supervisor: nil,
				Team:       []*Employee{},
			},
		},
	}

	empResult, err := empSchema.Parse(ceo)
	if err != nil {
		log.Fatal("Employee validation failed:", err)
	}

	fmt.Printf("✓ Employee validated: %s\n", empResult.Name)
	fmt.Printf("  Team members: %d\n", len(empResult.Team))

	fmt.Println()
	fmt.Println("Note: Circular references are automatically detected and handled")
	fmt.Println("using lazy evaluation to prevent stack overflow errors.")
}
