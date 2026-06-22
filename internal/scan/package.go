package scan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Package contains parsed Go package files.
type Package struct {
	Name  string
	Dir   string
	Files []*ast.File
	Fset  *token.FileSet
}

// ScanDir recursively parses Go packages under dir.
func ScanDir(dir string) ([]Package, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	var out []Package

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if shouldSkipDir(root, path) {
			return filepath.SkipDir
		}
		if !hasGoFiles(path) {
			return nil
		}

		pkgs, err := parser.ParseDir(fset, path, goSourceFileFilter, parser.ParseComments)
		if err != nil {
			return err
		}
		for name, pkg := range pkgs {
			files := make([]*ast.File, 0, len(pkg.Files))
			for _, file := range pkg.Files {
				files = append(files, file)
			}
			sort.Slice(files, func(i, j int) bool {
				return fset.Position(files[i].Package).Filename < fset.Position(files[j].Package).Filename
			})
			out = append(out, Package{
				Name:  name,
				Dir:   path,
				Files: files,
				Fset:  fset,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir == out[j].Dir {
			return out[i].Name < out[j].Name
		}
		return out[i].Dir < out[j].Dir
	})
	return out, nil
}

func shouldSkipDir(root, path string) bool {
	if path == root {
		return false
	}

	name := filepath.Base(path)
	switch name {
	case ".git", ".hg", ".svn", "vendor", "node_modules", "dist", "bin":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

func hasGoFiles(dir string) bool {
	entries, err := filepath.Glob(filepath.Join(dir, "*.go"))
	return err == nil && len(entries) > 0
}

func goSourceFileFilter(info fs.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && strings.HasSuffix(name, ".go")
}
