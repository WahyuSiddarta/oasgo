# oasgo

`oasgo` is a Go project for generating OpenAPI 3 YAML from Go source code.

## Goals

- Parse Go handlers, comments, types, and struct tags.
- Generate a valid OpenAPI 3 document.
- Emit deterministic YAML output.
- Work as both a library and, eventually, a CLI tool.
- Keep the generated API contract close to the Go code that implements it.

## Planned Capabilities

- Package and file scanning for Go projects.
- Route and operation extraction from annotations.
- Request and response schema generation from Go structs.
- Struct tag support for JSON field names and validation-related metadata.
- OpenAPI 3 components, paths, operations, parameters, request bodies, and responses.
- YAML output suitable for documentation portals, client generation, and contract review.

## Example Direction

Future usage may look like:

```go
// @Summary Create user
// @Description Creates a new user account.
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Create user payload"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// handler implementation
}
```

And generate an OpenAPI 3 YAML document:

```yaml
openapi: 3.0.3
info:
  title: Example API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
```

## Development Notes

This repository is currently a starting point. As implementation is added, keep parser logic, OpenAPI model construction, and YAML rendering separated so the project can support both CLI and library use cases.

Recommended implementation building blocks:

- `go/parser` and `go/ast` for syntax scanning.
- `go/types` for type resolution where needed.
- A structured OpenAPI model before YAML serialization.
- Fixture-based tests for generated YAML output.

## Status

Early project setup.
