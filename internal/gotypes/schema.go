package gotypes

// SchemaRef describes a Go type reference before it is converted into an
// OpenAPI schema.
type SchemaRef struct {
	Package string
	Name    string
}
