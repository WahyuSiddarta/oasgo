package openapi

// Schema describes an OpenAPI schema object.
type Schema struct {
	Ref                  string
	Type                 string
	Format               string
	Description          string
	Nullable             bool
	Items                *Schema
	Properties           map[string]*Schema
	Required             []string
	AdditionalProperties *Schema
}
