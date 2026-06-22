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

func TestRenderYAMLIncludesSupportedOperationObjects(t *testing.T) {
	doc := openapi.NewDocument("Example API", "1.0.0")
	doc.Components.Schemas["CreateUserRequest"] = &openapi.Schema{
		Type: "object",
		Properties: map[string]*openapi.Schema{
			"email": {Type: "string", Format: "email"},
			"name":  {Type: "string"},
		},
		Required: []string{"name", "email"},
	}
	doc.Components.Schemas["UserResponse"] = &openapi.Schema{
		Type: "object",
		Properties: map[string]*openapi.Schema{
			"id": {Type: "string"},
		},
		Required: []string{"id"},
	}
	doc.Paths["/users/{id}"] = &openapi.PathItem{Operations: map[string]*openapi.Operation{
		"post": {
			OperationID: "createUser",
			Summary:     "Create user",
			Tags:        []string{"users"},
			Parameters: []openapi.Parameter{
				{Name: "id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
			},
			RequestBody: &openapi.RequestBody{
				Required: true,
				Content: map[string]*openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/CreateUserRequest"}},
				},
			},
			Responses: map[string]*openapi.Response{
				"201": {
					Description: "User created.",
					Content: map[string]*openapi.MediaType{
						"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/UserResponse"}},
					},
				},
			},
		},
	}}

	got, err := RenderYAML(doc)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range [][]byte{
		[]byte("parameters:\n"),
		[]byte("requestBody:\n"),
		[]byte("content:\n"),
		[]byte("$ref: \"#/components/schemas/CreateUserRequest\"\n"),
		[]byte("components:\n"),
	} {
		if !bytes.Contains(got, want) {
			t.Fatalf("rendered YAML missing %q:\n%s", want, got)
		}
	}
}
