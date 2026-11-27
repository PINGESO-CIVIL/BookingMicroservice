package main

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestLayerImports(t *testing.T) {
	for _, layer := range []struct {
		path             string
		allowedInternal  []string
		forbiddenImports []string
	}{
		{"api", []string{"model", "service"}, []string{"database/sql", "github.com/jackc/pgx"}},
		{"service", []string{"model"}, []string{"net/http", "net/url", "database/sql", "github.com/jackc/pgx"}},
		{"store", []string{"model"}, []string{"net/http"}},
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
						pkg, _, _ := strings.Cut(internal, "/")
						if !slices.Contains(layer.allowedInternal, pkg) {
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
