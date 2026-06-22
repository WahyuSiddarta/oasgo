package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirRecursivelyParsesGoPackages(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package root\n")
	writeFile(t, root, "internal/api/handler.go", "package api\n")
	writeFile(t, root, "pkg/models/user.go", "package models\n")

	pkgs, err := ScanDir(root)
	if err != nil {
		t.Fatal(err)
	}

	got := packageNames(pkgs)
	want := []string{"root", "api", "models"}
	if len(got) != len(want) {
		t.Fatalf("packages = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("packages = %v, want %v", got, want)
		}
	}
}

func TestScanDirSkipsCommonNonSourceDirectories(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package root\n")
	writeFile(t, root, ".git/ignored.go", "package git\n")
	writeFile(t, root, "vendor/ignored.go", "package vendor\n")
	writeFile(t, root, "node_modules/ignored.go", "package node_modules\n")
	writeFile(t, root, "dist/ignored.go", "package dist\n")
	writeFile(t, root, "bin/ignored.go", "package bin\n")

	pkgs, err := ScanDir(root)
	if err != nil {
		t.Fatal(err)
	}

	got := packageNames(pkgs)
	if len(got) != 1 || got[0] != "root" {
		t.Fatalf("packages = %v, want [root]", got)
	}
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func packageNames(pkgs []Package) []string {
	names := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		names = append(names, pkg.Name)
	}
	return names
}
