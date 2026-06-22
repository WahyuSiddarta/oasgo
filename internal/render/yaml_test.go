package render

import (
	"bytes"
	"testing"

	"github.com/wahyusiddarta/oasgo/internal/openapi"
)

func TestRenderYAMLMinimalDocument(t *testing.T) {
	doc := openapi.NewDocument("Example API", "1.0.0")

	got, err := RenderYAML(doc)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("openapi: \"3.0.3\"\ninfo:\n  title: \"Example API\"\n  version: \"1.0.0\"\npaths: {}\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("rendered YAML mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderYAMLIsDeterministic(t *testing.T) {
	doc := openapi.NewDocument("Example API", "1.0.0")
	doc.Paths["/z"] = &openapi.PathItem{Operations: map[string]*openapi.Operation{
		"get": {Summary: "Z", Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
	}}
	doc.Paths["/a"] = &openapi.PathItem{Operations: map[string]*openapi.Operation{
		"post": {Summary: "A", Responses: map[string]*openapi.Response{"201": {Description: "Created"}}},
	}}

	first, err := RenderYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		next, err := RenderYAML(doc)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, next) {
			t.Fatalf("render is not deterministic\nfirst:\n%s\nnext:\n%s", first, next)
		}
	}
}
