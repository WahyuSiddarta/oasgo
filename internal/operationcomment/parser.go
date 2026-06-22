package operationcomment

import "errors"

// Operation is the typed representation of an oasgo operation comment.
type Operation struct {
	Route       Route
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	Request     *Request
	Responses   map[string]Response
}

// Route describes the HTTP method and OpenAPI path.
type Route struct {
	Method string
	Path   string
}

// Parameter describes a non-body operation parameter.
type Parameter struct {
	Name        string
	In          string
	Type        string
	Required    bool
	Description string
}

// Request describes a request body declaration.
type Request struct {
	ContentType string
	Body        *Body
}

// Body describes a typed request or response body.
type Body struct {
	Type        string
	Required    bool
	Description string
}

// Response describes a response declaration.
type Response struct {
	Description string
	Body        *Body
}

// Parse parses a raw oasgo operation comment block.
func Parse(block Block) (Operation, error) {
	if len(block.Lines) == 0 {
		return Operation{}, errors.New("oasgo: operation block is empty")
	}

	return Operation{}, errors.New("oasgo: operation comment parser is not implemented")
}
