package main

import (
	"go/parser"
	"go/token"
	"io/fs"
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
		allowHTTPSubdirs []string
	}{
		{"api", []string{"service", "service/demo"}, []string{"database/sql", "github.com/jackc/pgx"}, nil},
		{"service", []string{"service/demo", "service/domain"}, []string{"net/http", "net/url", "database/sql", "github.com/jackc/pgx"}, nil},
		{"store", []string{"repository", "service", "service/demo", "service/domain"}, []string{"net/http"}, []string{"notificaciones"}},
	} {
		t.Run(layer.path, func(t *testing.T) {
			layerRoot := filepath.Join("..", "..", "internal", layer.path)
			var files []string
			err := filepath.WalkDir(layerRoot, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil || len(files) == 0 {
				t.Fatalf("no se encontraron archivos Go: %v", err)
			}
			for _, name := range files {
				rel, err := filepath.Rel(layerRoot, name)
				if err != nil {
					t.Fatal(err)
				}
				topDir := strings.Split(filepath.ToSlash(rel), "/")[0]
				allowHTTP := slices.Contains(layer.allowHTTPSubdirs, topDir)

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
						if allowHTTP && forbidden == "net/http" {
							continue
						}
						if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
							t.Errorf("%s importa un paquete no permitido: %s", name, path)
						}
					}
				}
			}
		})
	}
}
