# oasgo

`oasgo` is a Go project for generating OpenAPI 3 YAML from Go source code.

The initial target is a practical subset of OpenAPI 3.0.3. Full OpenAPI Specification coverage is not implemented yet.

## Goals

- Parse Go handlers, comments, types, and struct tags.
- Generate valid OpenAPI 3.0.3 YAML for the supported feature subset.
- Emit deterministic YAML output.
- Work as a command line tool.
- Keep the generated API contract close to the Go code that implements it.

## Current Capabilities

- Recursive package and file scanning for Go projects.
- Route and operation extraction from declarative `oasgo:operation` comments.
- OpenAPI 3 paths, operations, parameters, request bodies, responses, and component schema references from parsed comments.
- Supported-subset OpenAPI 3.0.3 validation before YAML rendering.
- Deterministic YAML rendering for the current OpenAPI model.
- Local and release binary build commands.

## Planned Capabilities

- Request and response schema generation from Go structs.
- Struct tag support for JSON field names and validation-related metadata.
- Replacing placeholder component schemas with real schemas parsed from Go types.
- YAML output suitable for documentation portals, client generation, and contract review.

## OpenAPI Compliance Target

`oasgo` currently targets OpenAPI 3.0.3, not the full latest OpenAPI Specification. The first implementation should generate valid output for a focused subset:

- Root document fields: `openapi`, `info`, `paths`, and `components.schemas`.
- Path items and operations from `oasgo:operation` comments.
- Operation fields: `operationId`, `summary`, `description`, `tags`, `parameters`, `requestBody`, and `responses`.
- JSON request/response bodies using `application/json`.
- Component schemas generated from Go structs.
- Basic scalar, object, array, map, and referenced schemas.

Out of scope for the first version:

- Full OpenAPI 3.1.x support.
- Security schemes.
- Multiple content types per operation.
- Callbacks, links, webhooks, and advanced composition.
- Complete JSON Schema compatibility.
- Framework-specific route inference.

Generated output is validated against the current in-project OpenAPI 3.0.3 subset validator before YAML is rendered. External OpenAPI validator coverage should be added before claiming broader compliance.

## Getting Started

Install the CLI with `go install`:

```sh
go install github.com/wahyusiddarta/oasgo/cmd/oasgo@latest
```

Then run:

```sh
oasgo -dir ./examples/basic -title "Example API" -version "1.0.0"
```

During local development, run the command from this repository with `go run`:

```sh
go run ./cmd/oasgo -dir ./examples/basic -title "Example API" -version "1.0.0"
```

To build a local binary:

```sh
make build
./bin/oasgo -dir ./examples/basic -title "Example API" -version "1.0.0"
```

To build release binaries for macOS, Linux, and Windows:

```sh
make release VERSION=v0.1.0
```

Release binaries are written to `dist/`.

The command accepts:

- `-dir`: Go source directory tree to scan recursively. Defaults to the current directory.
- `-title`: OpenAPI `info.title`. Defaults to `API`.
- `-version`: OpenAPI `info.version`. Defaults to `0.0.0`.

Example output is written to stdout:

```sh
go run ./cmd/oasgo -dir ./examples/basic -title "Example API" -version "1.0.0" > openapi.yaml
```

Current status: the command layout, recursive Go package scanning, `oasgo:operation` parsing, deterministic YAML rendering, and supported-subset OpenAPI validation are in place. Go struct schema generation is still being implemented.

## Declarative Comments Format

`oasgo` uses declarative comments to describe HTTP operations next to the Go handler that implements them. A supported operation block starts with `oasgo:operation` inside a Go doc comment attached to a function or method.

No end marker is needed. The operation block ends when the attached Go doc comment ends.

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

This repository is currently a starting point. Parser logic, OpenAPI model construction, validation, and YAML rendering are kept separated behind the CLI.

See [docs/blueprint.md](docs/blueprint.md) for the implementation blueprint, package plan, milestone order, and testing strategy.
See [docs/declarative-comments.md](docs/declarative-comments.md) for the first supported source comment format.

Recommended implementation building blocks:

- `go/parser` and `go/ast` for syntax scanning.
- `go/types` for type resolution where needed.
- A structured OpenAPI model before YAML serialization.
- Validation before YAML serialization.
- Fixture-based tests for generated YAML output.

## Status

Early implementation. The CLI, recursive scanner, `oasgo:operation` parser, OpenAPI model, subset validator, renderer, build commands, and tests exist. The next major work is mapping Go structs into real `components.schemas` instead of placeholder object schemas.
