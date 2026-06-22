package openapi

// Paths maps OpenAPI paths to path items.
type Paths map[string]*PathItem

// PathItem groups operations by HTTP method.
type PathItem struct {
	Operations map[string]*Operation
}

// Operation describes one HTTP operation.
type Operation struct {
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]*Response
}

// Parameter describes an operation parameter.
type Parameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Schema      *Schema
}

// RequestBody describes an operation request body.
type RequestBody struct {
	Description string
	Required    bool
	Content     map[string]*MediaType
}

// Response describes an operation response.
type Response struct {
	Description string
	Content     map[string]*MediaType
}

// MediaType describes content for a request or response body.
type MediaType struct {
	Schema *Schema
}
