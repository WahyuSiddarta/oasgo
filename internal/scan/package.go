package scan

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// Package contains parsed Go package files.
type Package struct {
	Name  string
	Dir   string
	Files []*ast.File
	Fset  *token.FileSet
}

// ScanDir parses Go packages in dir.
func ScanDir(dir string) ([]Package, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	out := make([]Package, 0, len(pkgs))
	for name, pkg := range pkgs {
		files := make([]*ast.File, 0, len(pkg.Files))
		for _, file := range pkg.Files {
			files = append(files, file)
		}
		out = append(out, Package{
			Name:  name,
			Dir:   dir,
			Files: files,
			Fset:  fset,
		})
	}

	return out, nil
}
