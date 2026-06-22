# Declarative Comments Format

This document defines the first supported source comment format for `oasgo`.

The format is intentionally declarative: each handler declares the OpenAPI operation it represents, and `oasgo` maps the declaration plus referenced Go types into an OpenAPI 3 document.

## Design Goals

- Keep comments readable in normal Go source files.
- Avoid positional annotation formats where one missing field shifts the meaning of the rest of the line.
- Make parser errors precise and easy to fix.
- Keep the format namespaced so ordinary comments are ignored.
- Keep the first version small enough to implement and test thoroughly.
- Leave room for future OpenAPI fields without changing the basic shape.

## Operation Block

An operation block starts with `oasgo:operation` inside a Go doc comment attached to a function or method.

Example:

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

The parser strips the leading Go comment marker and parses the block body as a small YAML-like mapping. The block ends when the attached Go doc comment ends.

## Required Fields

Each operation block must include:

- `route`
- `responses`

Recommended fields:

- `operationId`
- `summary`
- `description`
- `tags`

If `operationId` is omitted, the generator may derive one from the Go function name, but generated IDs must remain deterministic.

## Field Reference

### `route`

Declares the HTTP method and path.

```yaml
route: POST /users/{id}
```

Rules:

- Method is required.
- Path is required.
- Method is case-insensitive in comments.
- Generated OpenAPI method keys are lower-case.
- Path parameters use OpenAPI `{name}` syntax.

### `operationId`

Declares the OpenAPI `operationId`.

```yaml
operationId: createUser
```

Rules:

- Must be unique across the generated document.
- Must be deterministic.
- If omitted, derive from the Go function or method name.

### `summary`

Short OpenAPI summary.

```yaml
summary: Create user
```

### `description`

Longer OpenAPI description.

```yaml
description: Creates a new user account.
```

For longer descriptions, use YAML block style:

```yaml
description: |
  Creates a new user account.
  Sends a welcome email after successful creation.
```

### `tags`

Declares operation tags.

```yaml
tags:
  - users
  - admin
```

Inline form is also allowed:

```yaml
tags: [users, admin]
```

### `parameters`

Declares non-body parameters.

```yaml
parameters:
  - name: id
    in: path
    type: string
    required: true
    description: User ID.
  - name: include_deleted
    in: query
    type: boolean
    required: false
    description: Include soft-deleted users.
```

Supported `in` values for version 1:

- `path`
- `query`
- `header`
- `cookie`

Rules:

- Path parameters are always required.
- Every `{name}` segment in `route` must have a matching path parameter.
- Every path parameter must appear in `route`.
- Parameter `type` can be a scalar type or a Go type name.

### `request`

Declares the request body.

```yaml
request:
  contentType: application/json
  body:
    type: CreateUserRequest
    required: true
    description: Create user payload.
```

Rules:

- Version 1 supports one request body.
- Default `contentType` is `application/json`.
- `body.type` references a Go type.
- `body.required` defaults to `false`.

### `responses`

Declares operation responses by HTTP status code.

```yaml
responses:
  "200":
    description: User found.
    body:
      type: UserResponse
  "404":
    description: User not found.
    body:
      type: ErrorResponse
```

Rules:

- At least one response is required.
- Status codes must be quoted in examples because YAML treats numeric keys differently across encoders.
- `description` is required for every response.
- `body` is optional for empty responses such as `204`.
- Default response content type is `application/json`.

Empty response example:

```yaml
responses:
  "204":
    description: User deleted.
```

## Type References

Type references use Go identifiers visible from the handler source file.

Examples:

```yaml
body:
  type: CreateUserRequest
```

```yaml
parameters:
  - name: filter
    in: query
    type: UserFilter
```

Supported first-version references:

- Built-in scalar names: `string`, `bool`, `int`, `int64`, `float64`.
- Local package type names.
- Imported package selector names such as `api.ErrorResponse`.
- Arrays using `[]Type`.

Generic instantiations, maps with non-string keys, and complex inline schemas are deferred.

## Struct Field Mapping

Go structs referenced by operation blocks become OpenAPI component schemas.

Rules:

- Use `json` tags for property names.
- Ignore fields tagged `json:"-"`.
- Fields with `omitempty` are optional.
- Fields without `omitempty` are required unless overridden later by validation support.
- Anonymous embedded structs are flattened only when doing so is unambiguous.
- Pointer nullability must be decided explicitly during schema implementation and then documented here.

Example:

```go
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}
```

Expected schema behavior:

- `name` and `email` are required.
- `age` is optional.

## Validation Tags

Validation tags are not part of the first parser milestone.

When added, the first supported constraints should be:

- `required`
- `min`
- `max`
- `minLength`
- `maxLength`
- `oneof`
- `email`

The implementation should document which Go validation tag formats are supported before parsing them.

## Parser Rules

The parser should:

- Only parse blocks that start with `oasgo:operation`.
- Ignore ordinary Go doc comments.
- Strip leading comment markers while preserving indentation inside the operation block.
- Report file and line for malformed blocks.
- Validate required fields before generation.
- Normalize HTTP methods and content types.
- Detect duplicate `operationId` values.
- Detect duplicate `route` method/path pairs.

## Error Examples

Missing route:

```text
handlers/user.go:12: oasgo operation requires route
```

Path parameter mismatch:

```text
handlers/user.go:17: route path parameter "id" has no matching parameter declaration
```

Duplicate operation:

```text
handlers/user.go:42: duplicate route GET /users/{id}; first declared at handlers/user.go:12
```

## Fixture Plan

Create fixtures around the comment format before implementing the full generator.

Recommended fixtures:

- `fixtures/simple`: one handler, one request struct, one response struct.
- `fixtures/path-query`: path and query parameters.
- `fixtures/empty-response`: `204` response without a body.
- `fixtures/multi-file`: handler and DTOs split across files.
- `fixtures/imported-types`: response type from another package.
- `fixtures/errors`: malformed operation blocks with expected parse errors.

## First Parser Acceptance Criteria

The first implementation of this format is acceptable when it can:

- Find `oasgo:operation` blocks on functions and methods.
- Parse route, operation ID, summary, description, tags, parameters, request body, and responses.
- Return typed records without generating OpenAPI yet.
- Report line-aware errors for malformed blocks.
- Pass unit tests and fixture tests for valid and invalid blocks.
