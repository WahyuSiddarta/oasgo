package openapi

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var pathParameterPattern = regexp.MustCompile(`\{([^}/]+)\}`)

// Validate checks whether doc satisfies the OpenAPI 3.0.3 subset supported by
// this project.
func Validate(doc *Document) error {
	var problems []string
	if doc == nil {
		return errors.New("oasgo: document is nil")
	}
	if doc.OpenAPI != Version {
		problems = append(problems, fmt.Sprintf("openapi must be %q", Version))
	}
	if strings.TrimSpace(doc.Info.Title) == "" {
		problems = append(problems, "info.title is required")
	}
	if strings.TrimSpace(doc.Info.Version) == "" {
		problems = append(problems, "info.version is required")
	}
	if doc.Paths == nil {
		problems = append(problems, "paths is required")
	}

	for _, path := range sortedKeys(doc.Paths) {
		item := doc.Paths[path]
		if item == nil {
			problems = append(problems, fmt.Sprintf("paths.%s must not be null", path))
			continue
		}
		if len(item.Operations) == 0 {
			problems = append(problems, fmt.Sprintf("paths.%s must define at least one operation", path))
			continue
		}

		for _, method := range sortedKeys(item.Operations) {
			location := fmt.Sprintf("paths.%s.%s", path, method)
			validateOperation(&problems, location, path, method, item.Operations[method], doc.Components.Schemas)
		}
	}
	for _, name := range sortedKeys(doc.Components.Schemas) {
		validateSchema(&problems, fmt.Sprintf("components.schemas.%s", name), doc.Components.Schemas[name], doc.Components.Schemas)
	}

	if len(problems) > 0 {
		return fmt.Errorf("oasgo: invalid OpenAPI document:\n- %s", strings.Join(problems, "\n- "))
	}

	return nil
}

func validateOperation(problems *[]string, location, path, method string, operation *Operation, schemas map[string]*Schema) {
	if operation == nil {
		*problems = append(*problems, fmt.Sprintf("%s must not be null", location))
		return
	}
	if !isHTTPMethod(method) {
		*problems = append(*problems, fmt.Sprintf("%s uses unsupported HTTP method %q", location, method))
	}
	if len(operation.Responses) == 0 {
		*problems = append(*problems, fmt.Sprintf("%s.responses is required", location))
	}

	routeParams := pathParameterNames(path)
	declaredPathParams := map[string]bool{}
	for i, parameter := range operation.Parameters {
		paramLocation := fmt.Sprintf("%s.parameters[%d]", location, i)
		validateParameter(problems, paramLocation, parameter, schemas)
		if parameter.In == "path" {
			declaredPathParams[parameter.Name] = true
			if !parameter.Required {
				*problems = append(*problems, fmt.Sprintf("%s.required must be true for path parameters", paramLocation))
			}
			if !routeParams[parameter.Name] {
				*problems = append(*problems, fmt.Sprintf("%s declares path parameter %q that is not present in route", paramLocation, parameter.Name))
			}
		}
	}
	for name := range routeParams {
		if !declaredPathParams[name] {
			*problems = append(*problems, fmt.Sprintf("%s is missing required path parameter %q", location, name))
		}
	}

	if operation.RequestBody != nil {
		validateRequestBody(problems, location+".requestBody", operation.RequestBody, schemas)
	}
	for _, status := range sortedKeys(operation.Responses) {
		validateResponse(problems, fmt.Sprintf("%s.responses.%s", location, status), operation.Responses[status], schemas)
	}
}

func validateParameter(problems *[]string, location string, parameter Parameter, schemas map[string]*Schema) {
	if strings.TrimSpace(parameter.Name) == "" {
		*problems = append(*problems, fmt.Sprintf("%s.name is required", location))
	}
	switch parameter.In {
	case "path", "query", "header", "cookie":
	default:
		*problems = append(*problems, fmt.Sprintf("%s.in must be one of path, query, header, cookie", location))
	}
	validateSchema(problems, location+".schema", parameter.Schema, schemas)
}

func validateRequestBody(problems *[]string, location string, requestBody *RequestBody, schemas map[string]*Schema) {
	if requestBody == nil {
		*problems = append(*problems, fmt.Sprintf("%s must not be null", location))
		return
	}
	if len(requestBody.Content) == 0 {
		*problems = append(*problems, fmt.Sprintf("%s.content is required", location))
	}
	for _, contentType := range sortedKeys(requestBody.Content) {
		validateMediaType(problems, fmt.Sprintf("%s.content.%s", location, contentType), requestBody.Content[contentType], schemas)
	}
}

func validateResponse(problems *[]string, location string, response *Response, schemas map[string]*Schema) {
	if response == nil {
		*problems = append(*problems, fmt.Sprintf("%s must not be null", location))
		return
	}
	if strings.TrimSpace(response.Description) == "" {
		*problems = append(*problems, fmt.Sprintf("%s.description is required", location))
	}
	for _, contentType := range sortedKeys(response.Content) {
		validateMediaType(problems, fmt.Sprintf("%s.content.%s", location, contentType), response.Content[contentType], schemas)
	}
}

func validateMediaType(problems *[]string, location string, mediaType *MediaType, schemas map[string]*Schema) {
	if mediaType == nil {
		*problems = append(*problems, fmt.Sprintf("%s must not be null", location))
		return
	}
	validateSchema(problems, location+".schema", mediaType.Schema, schemas)
}

func validateSchema(problems *[]string, location string, schema *Schema, schemas map[string]*Schema) {
	if schema == nil {
		*problems = append(*problems, fmt.Sprintf("%s is required", location))
		return
	}
	if schema.Ref != "" {
		name, ok := strings.CutPrefix(schema.Ref, "#/components/schemas/")
		if !ok || name == "" {
			*problems = append(*problems, fmt.Sprintf("%s.$ref must point to #/components/schemas/{name}", location))
			return
		}
		if _, exists := schemas[name]; !exists {
			*problems = append(*problems, fmt.Sprintf("%s.$ref points to missing schema %q", location, name))
		}
		return
	}
	if schema.Type == "" {
		*problems = append(*problems, fmt.Sprintf("%s.type is required when $ref is absent", location))
	}
	if schema.Type == "array" {
		validateSchema(problems, location+".items", schema.Items, schemas)
	}
	if schema.Type == "object" {
		for _, property := range sortedKeys(schema.Properties) {
			validateSchema(problems, fmt.Sprintf("%s.properties.%s", location, property), schema.Properties[property], schemas)
		}
		if schema.AdditionalProperties != nil {
			validateSchema(problems, location+".additionalProperties", schema.AdditionalProperties, schemas)
		}
	}
}

func isHTTPMethod(method string) bool {
	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}

func pathParameterNames(path string) map[string]bool {
	matches := pathParameterPattern.FindAllStringSubmatch(path, -1)
	names := make(map[string]bool, len(matches))
	for _, match := range matches {
		names[match[1]] = true
	}
	return names
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
