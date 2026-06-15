package gozod_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDocumentationDoesNotUseStalePublicLanguage(t *testing.T) {
	stale := regexp.MustCompile(`Uuid\(|Uuidv|FromGoZod|gozod\.Schema|Struct\(gozod\.ObjectSchema|New Feature|Maximum Performance`)
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

func documentationFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"README.md"}
	for _, dir := range []string{"docs", "SPECS"} {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || filepath.Ext(path) != ".md" {
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
