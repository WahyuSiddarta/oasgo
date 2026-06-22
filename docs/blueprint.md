# oasgo Blueprint

This document is the working blueprint for building `oasgo` into a command line tool that generates deterministic OpenAPI 3 YAML from Go source code.

The initial target is a practical OpenAPI 3.0.3 subset. Full OpenAPI Specification coverage is not a version 1 goal.

## Product Goal

`oasgo` should let a Go project keep its HTTP API contract close to the source code by scanning handlers, comments, request/response types, and struct tags, then emitting a reproducible OpenAPI 3 document.

Primary outcomes:

- Generate OpenAPI 3.0.3 YAML for the supported feature subset.
- Work as a command line tool first.
- Parse real Go syntax using the Go standard library.
- Keep output deterministic enough for reviewable diffs.
- Make examples and fixtures part of the public contract.

Non-goals for the first version:

- Runtime server instrumentation.
- Framework-specific reflection at process startup.
- Full OpenAPI feature coverage.
- OpenAPI 3.1.x support.
- Network-based validation or hosted documentation.

## Architecture

Keep the implementation split into small packages so scanning, interpretation, document construction, and rendering can evolve independently.

Suggested package layout:

```text
cmd/oasgo/
  main.go                  CLI entrypoint.

internal/scan/
  package.go               Loads packages and files.
  comment.go               Collects comments tied to declarations.

internal/operationcomment/
  parser.go                Parses `oasgo:operation` blocks into typed records.
  block.go                 Comment block extraction and validation.

internal/gotypes/
  resolver.go              Resolves Go identifiers and imports.
  schema.go                Maps Go types to schema descriptors.

internal/openapi/
  document.go              OpenAPI document model.
  operation.go             Path, operation, parameter, request, response model.
  schema.go                Component schema model.

internal/generate/
  generator.go             Coordinates scan, operation parsing, type mapping, and document construction.

internal/render/
  yaml.go                  Deterministic YAML rendering.

fixtures/
  simple/
  nested-types/
  tags/
  errors/
```

The Go packages under `internal/` are implementation details for the command. Keep their boundaries small and explicit so the CLI stays thin, but do not introduce a root public library API unless the project deliberately changes direction later.

## Data Flow

1. Scan Go packages from a configured root directory.
2. Collect functions, methods, comments, struct declarations, imports, and package metadata.
3. Parse supported `oasgo:operation` comment blocks into typed operation records.
4. Resolve referenced Go types through `go/ast`, `go/token`, and `go/types` where practical.
5. Build an OpenAPI document model from parsed operations and schemas.
6. Render YAML with stable key ordering and stable formatting.
7. Compare generated output against fixtures in tests.

## Declarative Comments Contract

The first design decision is the source comment format. Use [declarative-comments.md](declarative-comments.md) as the parser contract.

Version 1 uses a namespaced `oasgo:operation` block attached to Go functions or methods:

```go
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
```

Core decisions:

- Prefer declarative key/value blocks over positional comment directives.
- Keep the `oasgo:` namespace mandatory so ordinary Go comments are ignored.
- Parse into typed records before OpenAPI generation.
- Treat the comment format as a public contract once fixtures exist.
- Keep version 1 limited to route, operation ID, summary, description, tags, parameters, request body, and responses.

Struct field behavior:

- Prefer `json` tags for property names.
- Ignore fields tagged `json:"-"`.
- Treat `omitempty` as optional unless later validation rules say otherwise.
- Preserve declaration order where useful for readability, while rendering map keys deterministically.

Validation tag support should come after basic schema generation. When added, start with simple constraints such as `required`, `min`, `max`, `minLength`, `maxLength`, `oneof`, and `email`.

## OpenAPI Scope

Version 1 should support a valid OpenAPI 3.0.3 subset:

- `openapi`
- `info`
- `paths`
- `operationId`
- `summary`
- `description`
- `tags`
- `parameters`
- `requestBody`
- `responses`
- `components.schemas`
- Basic scalar schemas.
- Object schemas from structs.
- Array schemas from slices and arrays.
- Pointer unwrapping with nullable handling decided explicitly.

Version 1 compliance requirements:

- The root document must include `openapi`, `info`, and `paths`.
- Every operation must include at least one response.
- Every response must include `description`.
- Every path parameter in a route must have a matching required parameter.
- Every path parameter declaration must appear in the route.
- Request and response body content must render as valid OpenAPI `content`.
- Referenced schemas must exist in `components.schemas`.
- Renderer output must be deterministic.
- Fixture tests should include validator coverage before the project claims supported-subset compliance.

Defer until after the first stable generator:

- OpenAPI 3.1.x.
- Security schemes.
- Multiple content types.
- Deep object query parameters.
- Recursive types beyond safe reference emission.
- Generic type specialization.
- Polymorphism with `oneOf`, `anyOf`, and `allOf`.

## Implementation Plan

### Step 0: Declarative Comment Format

Finalize the `oasgo:operation` comment format before building parser or generator internals.

Deliverables:

- `docs/declarative-comments.md`.
- Fixture plan for valid and invalid comment blocks.
- A stable typed record shape for parsed operations.

Tests:

- No code tests are required for the document itself.
- The first parser milestone must implement tests directly from this contract.

### Step 1: Core OpenAPI Model

Build internal OpenAPI structs without scanning Go source yet.

Deliverables:

- `internal/openapi.Document`.
- Path, operation, parameter, request body, response, media type, and schema structs.
- Constructors or helpers only where they reduce repeated setup.

Tests:

- Unit tests for model defaults.
- Unit tests that verify required fields can represent a minimal OpenAPI 3 document.

### Step 2: Deterministic YAML Renderer

Render the OpenAPI model to YAML.

Deliverables:

- `internal/render.RenderYAML(doc *openapi.Document) ([]byte, error)`.
- Stable ordering for paths, methods, responses, parameters, and schemas.
- No random map iteration in generated output.

Tests:

- Golden file test for a minimal document.
- Golden file test with multiple paths and components to catch ordering drift.
- Idempotence test: repeated renders produce identical bytes.

### Step 3: Operation Comment Parser

Parse `oasgo:operation` comment blocks into typed operation records.

Deliverables:

- Typed records for route, operation ID, summary, description, tags, parameters, request body, and responses.
- Helpful parse errors with line context.
- Unit coverage for valid blocks and malformed blocks.

Tests:

- Valid operation blocks.
- Missing route.
- Invalid HTTP method.
- Invalid status code.
- Inline and block descriptions.

### Step 4: Source Scanner

Scan packages and collect function declarations with attached comments.

Deliverables:

- `internal/scan` package using `go/parser`, `go/ast`, and `go/token`.
- Records for files, package names, function declarations, comments, and source positions.
- A filter for generated/vendor/test files if needed.

Tests:

- Fixture package with one handler.
- Fixture package with multiple files.
- Fixture package with method receivers.
- Fixture package with unrelated comments that should be ignored.

### Step 5: Type Resolver and Schema Mapper

Map Go request/response types to OpenAPI component schemas.

Deliverables:

- Scalar mapping for `string`, `bool`, signed/unsigned ints, floats, and time-like types.
- Struct schema generation using JSON tags.
- Slice, array, pointer, and embedded struct support.
- Component naming strategy with collision handling.

Tests:

- Scalar fields.
- JSON tag renaming.
- `json:"-"` omission.
- Optional fields from `omitempty`.
- Nested structs and slices.
- Imported package types.

### Step 6: Operation Generator

Combine parsed operation comments and schema mapping into OpenAPI paths.

Deliverables:

- Generate path operations from `route`.
- Generate path/query/header/cookie parameters from `parameters`.
- Generate request bodies from body params.
- Generate responses from `responses`.
- Register component schemas once and reuse references.

Tests:

- One route with request body and response body.
- Multiple HTTP methods on the same path.
- Path parameter extraction.
- Query parameter extraction.
- Error responses.
- Duplicate route/method conflict handling.

### Step 7: Library API

Expose a small public API.

Deliverables:

- Root package functions for document and YAML generation.
- Config validation.
- Error wrapping that points users to source files and lines.

Tests:

- Library generation from a fixture directory.
- Config validation errors.
- Stable output through the public API.

### Step 8: CLI

Expand the CLI over the internal generator.

Deliverables:

- `cmd/oasgo`.
- Flags for directory, output file, title, and version.
- Default output to stdout when no file is supplied.
- Exit codes suitable for CI.

Tests:

- CLI help output.
- CLI generation from fixture to stdout.
- CLI generation to a file.
- Invalid config returns a non-zero status.

### Step 9: Documentation and Examples

Turn fixtures into useful examples.

Deliverables:

- README quickstart.
- Example handler operation comments.
- Example generated OpenAPI YAML.
- Notes on supported comment fields and current limitations.

Tests:

- Keep README examples aligned with fixtures where practical.
- Fixture golden output remains the contract for docs examples.

## Testing Strategy

Use focused unit tests for parser and type behavior, then fixture-based tests for generated output.

Test layers:

- Operation comment parser unit tests.
- Scanner unit tests using tiny in-repo fixture packages.
- Schema mapper unit tests using Go structs from fixtures.
- Generator tests that compare complete OpenAPI output.
- Renderer golden tests for deterministic YAML.
- CLI tests only after the internal generator behavior is stable.

Recommended command:

```sh
/Volumes/KyoMac/go/bin/go test ./...
```

For local cache issues, use a writable cache:

```sh
GOCACHE=/private/tmp/oasgo-gocache /Volumes/KyoMac/go/bin/go test ./...
```

Golden test rules:

- Store expected YAML under `fixtures/<case>/openapi.yaml`.
- Review fixture diffs like API contract changes.
- Avoid updating golden files without understanding the behavior change.

## Determinism Rules

Generated output must be reproducible.

Rules:

- Sort map-derived keys before rendering.
- Preserve explicit source order only where it improves readability and is stable.
- Avoid timestamps, absolute local paths, random IDs, and environment-dependent values in output.
- Normalize HTTP methods to lower-case OpenAPI path method keys.
- Normalize content type aliases to canonical values such as `application/json`.

## Error Handling

Errors should help users fix operation comments quickly.

Good error shape:

```text
handler.go:18: response "400" requires description
```

Guidelines:

- Include file and line when source context exists.
- Distinguish parse errors, type-resolution errors, route conflicts, and render errors.
- Prefer collecting independent operation comment errors in one pass if it does not make control flow brittle.

## Release Checklist

Before a first usable release:

- CLI can generate YAML from at least one fixture project.
- Renderer is deterministic.
- README has a quickstart.
- Supported comment fields are documented.
- Fixture output is committed.
- `go test ./...` passes locally.

## Near-Term Milestones

1. Add the internal OpenAPI model and renderer.
2. Add operation comment parsing with parser unit tests.
3. Add scanner fixtures and source discovery.
4. Add struct schema generation.
5. Generate one complete route into OpenAPI YAML.
6. Add the first CLI command.
7. Expand fixtures based on real usage.
