package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/wahyusiddarta/oasgo/internal/generate"
	"github.com/wahyusiddarta/oasgo/internal/render"
)

func main() {
	dir := flag.String("dir", ".", "Go source directory to scan")
	title := flag.String("title", "API", "OpenAPI info.title")
	version := flag.String("version", "0.0.0", "OpenAPI info.version")
	flag.Parse()

	doc, err := generate.Generate(context.Background(), generate.Config{
		Dir:     *dir,
		Title:   *title,
		Version: *version,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, err := render.RenderYAML(doc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(string(out))
}
