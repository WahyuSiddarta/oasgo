package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wahyusiddarta/oasgo/internal/render"
)

func TestGenerateBuildsOperationsFromComments(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "internal/api/handler.go", `package api

// CreateUser creates a new user.
//
// oasgo:operation
// route: POST /users
// operationId: createUser
// summary: Create user
// tags:
//   - users
// request:
//   body:
//     type: CreateUserRequest
//     required: true
// responses:
//   "201":
//     description: User created.
//     body:
//       type: UserResponse
func CreateUser() {}
`)

	doc, err := Generate(context.Background(), Config{
		Dir:     root,
		Title:   "Example API",
		Version: "1.0.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	if doc.Paths["/users"] == nil || doc.Paths["/users"].Operations["post"] == nil {
		t.Fatalf("paths = %#v", doc.Paths)
	}
	if doc.Components.Schemas["CreateUserRequest"] == nil {
		t.Fatal("missing CreateUserRequest placeholder schema")
	}
	if doc.Components.Schemas["UserResponse"] == nil {
		t.Fatal("missing UserResponse placeholder schema")
	}
	if _, err := render.RenderYAML(doc); err != nil {
		t.Fatal(err)
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
