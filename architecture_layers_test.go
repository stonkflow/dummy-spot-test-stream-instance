package main

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	layerDomain    = "domain"
	layerUsecase   = "usecase"
	layerCodec     = "codec"
	layerTransport = "transport"
	layerApp       = "app"
)

func TestLayerDependencies(t *testing.T) {
	modulePath := readModulePath(t)
	internalPrefix := modulePath + "/internal/"

	err := filepath.WalkDir("internal", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		srcLayer := layerFromFile(path)
		if srcLayer == "" {
			return nil
		}

		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		for _, imp := range file.Imports {
			impPath := strings.Trim(imp.Path.Value, "\"")
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			dstLayer := layerFromImportPath(internalPrefix, impPath)
			if dstLayer == "" {
				continue
			}

			if isForbiddenDependency(srcLayer, dstLayer) {
				t.Errorf("forbidden dependency: %s (%s) -> %s (%s)", path, srcLayer, impPath, dstLayer)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk internal: %v", err)
	}
}

func readModulePath(t *testing.T) string {
	t.Helper()

	file, err := os.Open("go.mod")
	if err != nil {
		t.Fatalf("open go.mod: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	if err = scanner.Err(); err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	t.Fatal("module path not found in go.mod")
	return ""
}

func layerFromFile(path string) string {
	clean := filepath.ToSlash(path)
	switch {
	case strings.HasPrefix(clean, "internal/domain/"):
		return layerDomain
	case strings.HasPrefix(clean, "internal/usecase/"):
		return layerUsecase
	case strings.HasPrefix(clean, "internal/codec/"):
		return layerCodec
	case strings.HasPrefix(clean, "internal/transport/"):
		return layerTransport
	case strings.HasPrefix(clean, "internal/app/"):
		return layerApp
	default:
		return ""
	}
}

func layerFromImportPath(internalPrefix, imp string) string {
	rel := strings.TrimPrefix(imp, internalPrefix)
	switch {
	case strings.HasPrefix(rel, "domain"):
		return layerDomain
	case strings.HasPrefix(rel, "usecase"):
		return layerUsecase
	case strings.HasPrefix(rel, "codec"):
		return layerCodec
	case strings.HasPrefix(rel, "transport"):
		return layerTransport
	case strings.HasPrefix(rel, "app"):
		return layerApp
	default:
		return ""
	}
}

func isForbiddenDependency(src, dst string) bool {
	switch src {
	case layerDomain:
		return dst != layerDomain
	case layerUsecase:
		return dst == layerTransport || dst == layerApp || dst == layerCodec
	case layerCodec:
		return dst != layerDomain
	case layerTransport:
		return dst == layerApp || dst == layerDomain
	case layerApp:
		return dst == layerTransport || dst == layerCodec || dst == layerDomain
	default:
		return false
	}
}
