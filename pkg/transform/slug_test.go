package transform

import (
	"fmt"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Zod v4 reference test cases
		{name: "basic words", input: "Hello World", want: "hello-world"},
		{name: "padded spaces", input: "  Hello   World  ", want: "hello-world"},
		{name: "special chars removed", input: "Hello@World#123", want: "helloworld123"},
		{name: "preserves hyphens", input: "Hello-World", want: "hello-world"},
		{name: "underscores to hyphens", input: "Hello_World", want: "hello-world"},
		{name: "collapses hyphens", input: "---Hello---World---", want: "hello-world"},
		{name: "collapses spaces", input: "Hello  World", want: "hello-world"},
		{name: "strips all special", input: "Hello!@#$%^&*()World", want: "helloworld"},

		// Additional edge cases
		{name: "empty", input: "", want: ""},
		{name: "only special chars", input: "!@#$%^&*()", want: ""},
		{name: "only spaces", input: "   ", want: ""},
		{name: "mixed delimiters", input: "a_b-c d", want: "a-b-c-d"},
		{name: "numbers only", input: "123", want: "123"},
		{name: "leading trailing hyphens", input: "-Leading-Trailing-", want: "leading-trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.input); got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkSlugify(b *testing.B) {
	inputs := []string{
		"Hello World",
		"Special!@#$%^&*()Characters",
		"Multiple   Spaces   Here",
		"Under_Score_And-Hyphens",
		"ThisIsAVeryLongStringWithManyCharactersThatNeedsToBeSlugified",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = Slugify(s)
		}
	}
}

func BenchmarkSlugifyShort(b *testing.B) {
	for b.Loop() {
		_ = Slugify("Hello World")
	}
}

func BenchmarkSlugifyLong(b *testing.B) {
	s := "This Is A Very Long String With Many Words That Needs To Be Slugified For URL Usage"
	for b.Loop() {
		_ = Slugify(s)
	}
}

func BenchmarkSlugifyComplex(b *testing.B) {
	s := "Special!@#$%^&*()Characters___And---Multiple---Delimiters"
	for b.Loop() {
		_ = Slugify(s)
	}
}

func ExampleSlugify() {
	fmt.Println(Slugify("Hello World"))
	fmt.Println(Slugify("  Multiple   Spaces  "))
	fmt.Println(Slugify("Under_Score_And-Hyphens"))
	// Output:
	// hello-world
	// multiple-spaces
	// under-score-and-hyphens
}
