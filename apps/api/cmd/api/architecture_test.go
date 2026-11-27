package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestInternalLayout(t *testing.T) {
	const root = "../../internal"
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}
	want := []string{"api", "repository", "service", "store"}
	if !slices.Equal(directories, want) {
		t.Fatalf("directorios de internal: %v; plantilla: %v", directories, want)
	}
}

func TestLayerImports(t *testing.T) {
	for _, layer := range []struct {
		path             string
		allowedInternal  []string
		forbiddenImports []string
	}{
		{"api", []string{"service", "service/demo"}, []string{"database/sql", "github.com/jackc/pgx"}},
		{"service", []string{"service/demo", "service/domain"}, []string{"net/http", "net/url", "database/sql", "github.com/jackc/pgx"}},
		{"store", []string{"service/demo", "service/domain"}, []string{"net/http"}},
	} {
		t.Run(layer.path, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", "..", "internal", layer.path, "*.go"))
			if err != nil || len(files) == 0 {
				t.Fatalf("no se encontraron archivos Go: %v", err)
			}
			for _, name := range files {
				if strings.HasSuffix(name, "_test.go") {
					continue
				}
				file, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
				if err != nil {
					t.Fatal(err)
				}
				for _, spec := range file.Imports {
					path, err := strconv.Unquote(spec.Path.Value)
					if err != nil {
						t.Fatal(err)
					}
					if internal, ok := strings.CutPrefix(path, "bookingms/internal/"); ok {
						if !slices.Contains(layer.allowedInternal, internal) {
							t.Errorf("%s importa una capa no permitida: %s", name, path)
						}
					}
					for _, forbidden := range layer.forbiddenImports {
						if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
							t.Errorf("%s importa un paquete no permitido: %s", name, path)
						}
					}
				}
			}
		})
	}
}
