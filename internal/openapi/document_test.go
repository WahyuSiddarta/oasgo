package openapi

import "testing"

func TestNewDocumentDefaults(t *testing.T) {
	doc := NewDocument("Example API", "1.0.0")

	if doc.OpenAPI != "3.0.3" {
		t.Fatalf("OpenAPI version = %q, want %q", doc.OpenAPI, "3.0.3")
	}
	if doc.Info.Title != "Example API" {
		t.Fatalf("title = %q, want %q", doc.Info.Title, "Example API")
	}
	if doc.Info.Version != "1.0.0" {
		t.Fatalf("version = %q, want %q", doc.Info.Version, "1.0.0")
	}
	if doc.Paths == nil {
		t.Fatal("Paths is nil")
	}
	if doc.Components.Schemas == nil {
		t.Fatal("Components.Schemas is nil")
	}
}
