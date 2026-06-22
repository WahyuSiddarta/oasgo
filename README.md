# oasgo

`oasgo` is a Go project for generating OpenAPI 3 YAML from Go source code.

## Goals

- Parse Go handlers, comments, types, and struct tags.
- Generate a valid OpenAPI 3 document.
- Emit deterministic YAML output.
- Work as a command line tool.
- Keep the generated API contract close to the Go code that implements it.

## Planned Capabilities

- Package and file scanning for Go projects.
- Route and operation extraction from declarative `oasgo:operation` comments.
- Request and response schema generation from Go structs.
- Struct tag support for JSON field names and validation-related metadata.
- OpenAPI 3 components, paths, operations, parameters, request bodies, and responses.
- YAML output suitable for documentation portals, client generation, and contract review.

## Getting Started

Run the command from this repository with `go run`:

```sh
go run ./cmd/oasgo -dir ./examples/basic -title "Example API" -version "1.0.0"
```

The command accepts:

- `-dir`: Go source directory to scan. Defaults to the current directory.
- `-title`: OpenAPI `info.title`. Defaults to `API`.
- `-version`: OpenAPI `info.version`. Defaults to `0.0.0`.

Example output is written to stdout:

```sh
go run ./cmd/oasgo -dir ./examples/basic -title "Example API" -version "1.0.0" > openapi.yaml
```

Current status: the command and package layout are in place, but operation parsing and schema generation are still being implemented. At this stage, the command emits the initial OpenAPI document shape.

## Declarative Comments Format

`oasgo` uses declarative comments to describe HTTP operations next to the Go handler that implements them. A supported operation block starts with `oasgo:operation` inside a Go doc comment attached to a function or method.

The comment body is a small YAML-like mapping. The required fields are:

- `route`: HTTP method and OpenAPI path, for example `POST /users` or `GET /users/{id}`.
- `responses`: response definitions keyed by HTTP status code.

Recommended fields are:

- `operationId`: stable OpenAPI operation ID.
- `summary`: short operation summary.
- `description`: longer operation description.
- `tags`: OpenAPI operation tags.

Supported sections planned for the first implementation are:

- `parameters`: path, query, header, and cookie parameters.
- `request`: one request body, defaulting to `application/json`.
- `responses`: one or more responses, each with a required description and optional body.

## Example Direction

Future usage may look like:

```go
// CreateUser creates a new user account.
//
// oasgo:operation
// route: POST /users
// operationId: createUser
// summary: Create user
// description: Creates a new user account.
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
//   "400":
//     description: Invalid request.
//     body:
//       type: ErrorResponse
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// handler implementation
}
```

Important format rules:

- Method names in `route` are case-insensitive, but generated OpenAPI method keys are lower-case.
- Path parameters use OpenAPI `{name}` syntax and must have matching `parameters` entries.
- Response status codes should be quoted in examples, such as `"201"`.
- `request.body.type` and `responses.*.body.type` reference Go type names.
- `json` struct tags will drive schema property names once schema generation is implemented.

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

This repository is currently a starting point. As implementation is added, keep parser logic, OpenAPI model construction, and YAML rendering separated behind the CLI.

See [docs/blueprint.md](docs/blueprint.md) for the implementation blueprint, package plan, milestone order, and testing strategy.
See [docs/declarative-comments.md](docs/declarative-comments.md) for the first supported source comment format.

Recommended implementation building blocks:

- `go/parser` and `go/ast` for syntax scanning.
- `go/types` for type resolution where needed.
- A structured OpenAPI model before YAML serialization.
- Fixture-based tests for generated YAML output.

## Status

Early project setup.
