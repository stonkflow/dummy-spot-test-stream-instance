package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalPackagesHaveDocGo(t *testing.T) {
	err := filepath.WalkDir("internal", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}

		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		hasProdGo := false
		hasDocGo := false

		for _, file := range files {
			name := file.Name()
			if file.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			hasProdGo = true
			if name == "doc.go" {
				hasDocGo = true
			}
		}

		if !hasProdGo {
			return nil
		}
		if !hasDocGo {
			t.Errorf("package %s must contain doc.go", path)
			return nil
		}

		docPath := filepath.Join(path, "doc.go")
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, docPath, nil, parser.ParseComments)
		if err != nil {
			t.Errorf("parse %s: %v", docPath, err)
			return nil
		}
		if parsed.Doc == nil {
			t.Errorf("doc.go in %s must contain a package comment", path)
			return nil
		}

		comment := strings.TrimSpace(parsed.Doc.Text())
		if !strings.HasPrefix(comment, "Package ") {
			t.Errorf("package comment in %s/doc.go should start with 'Package '", path)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("walk internal: %v", err)
	}
}
