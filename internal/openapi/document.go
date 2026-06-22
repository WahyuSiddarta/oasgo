package openapi

const Version = "3.0.3"

// Document is the root OpenAPI 3 document model.
type Document struct {
	OpenAPI    string
	Info       Info
	Paths      Paths
	Components Components
}

// Info describes the generated API document.
type Info struct {
	Title   string
	Version string
}

// Components contains reusable OpenAPI document parts.
type Components struct {
	Schemas map[string]*Schema
}

// NewDocument creates a document with stable defaults.
func NewDocument(title, version string) *Document {
	return &Document{
		OpenAPI: Version,
		Info: Info{
			Title:   title,
			Version: version,
		},
		Paths: Paths{},
		Components: Components{
			Schemas: map[string]*Schema{},
		},
	}
}
