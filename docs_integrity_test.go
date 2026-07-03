package gozod_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kaptinlin/gozod"
)

func TestDocumentationDoesNotUseStalePublicLanguage(t *testing.T) {
	stale := regexp.MustCompile(`Uuid\(|Uuidv|FromGoZod|gozod\.Schema|Struct\(gozod\.ObjectSchema|New Feature|Maximum Performance|registry\.Add\(\s*(?:schema|[[:alpha:]_][[:alnum:]_]*Schema)\s*\)|FromStruct.*auto|auto.*FromStruct|automatically detects and uses generated|Uses generated code automatically|Automatic detection by FromStruct|gozod\.Config\([^)]|GetLocaleFormatter|5-10x faster|50-70%|3-5x`)
	for _, path := range documentationFiles(t) {
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			if stale.MatchString(scanner.Text()) {
				t.Errorf("%s:%d uses stale public language: %s", path, line, scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("scan %s: %v", path, err)
		}
		if err := file.Close(); err != nil {
			t.Errorf("close %s: %v", path, err)
		}
	}
}

func TestDocumentationJSONSchemaRegistryExampleCompiles(t *testing.T) {
	registry := gozod.NewRegistry[gozod.GlobalMeta]()

	var userSchema, postSchema gozod.ZodSchema
	userSchema = gozod.Object(gozod.ObjectSchema{
		"id":   gozod.UUID(),
		"name": gozod.String(),
		"posts": gozod.Lazy(func() gozod.ZodSchema {
			return gozod.Slice[any](postSchema)
		}),
	})
	postSchema = gozod.Object(gozod.ObjectSchema{
		"id":     gozod.UUID(),
		"title":  gozod.String(),
		"author": gozod.Lazy(func() gozod.ZodSchema { return userSchema }),
	})

	registry.Add(userSchema, gozod.GlobalMeta{ID: "User"})
	registry.Add(postSchema, gozod.GlobalMeta{ID: "Post"})

	if _, err := gozod.ToJSONSchema(registry); err != nil {
		t.Fatalf("ToJSONSchema registry example: %v", err)
	}
}

func documentationFiles(t *testing.T) []string {
	t.Helper()
	files := existingDocumentationFiles("README.md", "AGENTS.md", "CLAUDE.md")
	for _, dir := range []string{"docs", "SPECS", "examples"} {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			switch filepath.Ext(path) {
			case ".md", ".go":
			default:
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	return files
}

func existingDocumentationFiles(paths ...string) []string {
	files := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	return files
}
