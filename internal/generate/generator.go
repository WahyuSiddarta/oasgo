package generate

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/wahyusiddarta/oasgo/internal/openapi"
	"github.com/wahyusiddarta/oasgo/internal/operationcomment"
	"github.com/wahyusiddarta/oasgo/internal/scan"
)

// Config controls generator behavior.
type Config struct {
	Dir     string
	Title   string
	Version string
}

// Generate coordinates source scanning, operation parsing, schema generation,
// and OpenAPI document construction.
func Generate(ctx context.Context, cfg Config) (*openapi.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cfg.Dir == "" {
		return nil, errors.New("oasgo: Dir is required")
	}

	pkgs, err := scan.ScanDir(cfg.Dir)
	if err != nil {
		return nil, err
	}

	doc := openapi.NewDocument(cfg.Title, cfg.Version)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				lines := scan.CommentLines(fn.Doc)
				block, ok := operationcomment.ExtractBlock(lines)
				if !ok {
					continue
				}
				operation, err := operationcomment.Parse(block)
				if err != nil {
					return nil, fmt.Errorf("%s: %s: %w", pkg.Dir, fn.Name.Name, err)
				}
				addOperation(doc, operation)
			}
		}
	}
	return doc, nil
}

func addOperation(doc *openapi.Document, operation operationcomment.Operation) {
	ensureReferencedSchemas(doc, operation)

	path := operation.Route.Path
	method := strings.ToLower(operation.Route.Method)
	if doc.Paths[path] == nil {
		doc.Paths[path] = &openapi.PathItem{Operations: map[string]*openapi.Operation{}}
	}

	doc.Paths[path].Operations[method] = &openapi.Operation{
		OperationID: operation.OperationID,
		Summary:     operation.Summary,
		Description: operation.Description,
		Tags:        operation.Tags,
		Parameters:  mapParameters(operation.Parameters),
		RequestBody: mapRequest(operation.Request),
		Responses:   mapResponses(operation.Responses),
	}
}

func ensureReferencedSchemas(doc *openapi.Document, operation operationcomment.Operation) {
	for _, parameter := range operation.Parameters {
		ensureSchema(doc, parameter.Type)
	}
	if operation.Request != nil && operation.Request.Body != nil {
		ensureSchema(doc, operation.Request.Body.Type)
	}
	for _, response := range operation.Responses {
		if response.Body != nil {
			ensureSchema(doc, response.Body.Type)
		}
	}
}

func ensureSchema(doc *openapi.Document, typeName string) {
	if isScalarType(typeName) {
		return
	}
	if doc.Components.Schemas[typeName] == nil {
		doc.Components.Schemas[typeName] = &openapi.Schema{Type: "object"}
	}
}

func mapParameters(parameters []operationcomment.Parameter) []openapi.Parameter {
	out := make([]openapi.Parameter, 0, len(parameters))
	for _, parameter := range parameters {
		out = append(out, openapi.Parameter{
			Name:        parameter.Name,
			In:          parameter.In,
			Description: parameter.Description,
			Required:    parameter.Required,
			Schema:      schemaFromType(parameter.Type),
		})
	}
	return out
}

func mapRequest(request *operationcomment.Request) *openapi.RequestBody {
	if request == nil || request.Body == nil {
		return nil
	}
	contentType := request.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	return &openapi.RequestBody{
		Description: request.Body.Description,
		Required:    request.Body.Required,
		Content: map[string]*openapi.MediaType{
			contentType: {Schema: schemaFromType(request.Body.Type)},
		},
	}
}

func mapResponses(responses map[string]operationcomment.Response) map[string]*openapi.Response {
	out := make(map[string]*openapi.Response, len(responses))
	for status, response := range responses {
		out[status] = &openapi.Response{
			Description: response.Description,
			Content:     mapBodyContent(response.Body),
		}
	}
	return out
}

func mapBodyContent(body *operationcomment.Body) map[string]*openapi.MediaType {
	if body == nil || body.Type == "" {
		return nil
	}
	return map[string]*openapi.MediaType{
		"application/json": {Schema: schemaFromType(body.Type)},
	}
}

func schemaFromType(typeName string) *openapi.Schema {
	switch typeName {
	case "string":
		return &openapi.Schema{Type: "string"}
	case "integer", "int", "int8", "int16", "int32", "int64":
		return &openapi.Schema{Type: "integer"}
	case "number", "float32", "float64":
		return &openapi.Schema{Type: "number"}
	case "boolean", "bool":
		return &openapi.Schema{Type: "boolean"}
	case "":
		return &openapi.Schema{Type: "object"}
	default:
		return &openapi.Schema{Ref: "#/components/schemas/" + typeName}
	}
}

func isScalarType(typeName string) bool {
	switch typeName {
	case "", "string", "integer", "int", "int8", "int16", "int32", "int64", "number", "float32", "float64", "boolean", "bool":
		return true
	default:
		return false
	}
}
