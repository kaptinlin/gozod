package validate

import (
	"testing"
)

// BenchmarkSlugify tests the performance of the Slugify function
func BenchmarkSlugify(b *testing.B) {
	testCases := []string{
		"Hello World",
		"Special!@#$%^&*()Characters",
		"Multiple   Spaces   Here",
		"Under_Score_And-Hyphens",
		"ThisIsAVeryLongStringWithManyCharactersThatNeedsToBeSlugified",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = Slugify(tc)
		}
	}
}

// BenchmarkSlugifyShort benchmarks short strings
func BenchmarkSlugifyShort(b *testing.B) {
	input := "Hello World"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Slugify(input)
	}
}

// BenchmarkSlugifyLong benchmarks long strings
func BenchmarkSlugifyLong(b *testing.B) {
	input := "This Is A Very Long String With Many Words That Needs To Be Slugified For URL Usage"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Slugify(input)
	}
}

// BenchmarkSlugifyComplex benchmarks strings with special characters
func BenchmarkSlugifyComplex(b *testing.B) {
	input := "Special!@#$%^&*()Characters___And---Multiple---Delimiters"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Slugify(input)
	}
}

// TestSlugifyCorrectness ensures optimization doesn't break functionality
func TestSlugifyCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "Trim spaces",
			input:    "  Trim  Me  ",
			expected: "trim-me",
		},
		{
			name:     "Special chars",
			input:    "Special!@#$%Chars",
			expected: "specialchars",
		},
		{
			name:     "Underscores",
			input:    "Under_Score",
			expected: "under-score",
		},
		{
			name:     "Multiple hyphens",
			input:    "Multiple---Hyphens",
			expected: "multiple-hyphens",
		},
		{
			name:     "Mixed",
			input:    "Mix123Numbers",
			expected: "mix123numbers",
		},
		{
			name:     "Leading trailing hyphens",
			input:    "-Leading-Trailing-",
			expected: "leading-trailing",
		},
		{
			name:     "Empty",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special chars",
			input:    "!@#$%^&*()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
