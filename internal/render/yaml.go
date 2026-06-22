package render

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/wahyusiddarta/oasgo/internal/openapi"
)

// RenderYAML renders an OpenAPI document with deterministic field ordering.
func RenderYAML(doc *openapi.Document) ([]byte, error) {
	if err := openapi.Validate(doc); err != nil {
		return nil, err
	}

	var b bytes.Buffer
	writeKV(&b, 0, "openapi", doc.OpenAPI)
	b.WriteString("info:\n")
	writeKV(&b, 2, "title", doc.Info.Title)
	writeKV(&b, 2, "version", doc.Info.Version)

	b.WriteString("paths:")
	if len(doc.Paths) == 0 {
		b.WriteString(" {}\n")
	} else {
		b.WriteByte('\n')
		for _, path := range sortedKeys(doc.Paths) {
			fmt.Fprintf(&b, "  %s:\n", quoteString(path))
			item := doc.Paths[path]
			for _, method := range sortedKeys(item.Operations) {
				fmt.Fprintf(&b, "    %s:\n", method)
				writeOperation(&b, 6, item.Operations[method])
			}
		}
	}

	if len(doc.Components.Schemas) > 0 {
		b.WriteString("components:\n")
		b.WriteString("  schemas:\n")
		for _, name := range sortedKeys(doc.Components.Schemas) {
			fmt.Fprintf(&b, "    %s:\n", name)
			writeSchema(&b, 6, doc.Components.Schemas[name])
		}
	}

	return b.Bytes(), nil
}

func writeOperation(b *bytes.Buffer, indent int, operation *openapi.Operation) {
	writeOptionalKV(b, indent, "operationId", operation.OperationID)
	writeOptionalKV(b, indent, "summary", operation.Summary)
	writeOptionalKV(b, indent, "description", operation.Description)
	writeStringSlice(b, indent, "tags", operation.Tags)
	writeParameters(b, indent, operation.Parameters)

	if operation.RequestBody != nil {
		writeIndent(b, indent)
		b.WriteString("requestBody:\n")
		writeOptionalKV(b, indent+2, "description", operation.RequestBody.Description)
		if operation.RequestBody.Required {
			writeKV(b, indent+2, "required", "true")
		}
		writeContent(b, indent+2, operation.RequestBody.Content)
	}

	writeResponses(b, indent, operation.Responses)
}

func writeParameters(b *bytes.Buffer, indent int, parameters []openapi.Parameter) {
	if len(parameters) == 0 {
		return
	}
	writeIndent(b, indent)
	b.WriteString("parameters:\n")
	for _, parameter := range parameters {
		writeIndent(b, indent+2)
		fmt.Fprintf(b, "- name: %s\n", quoteString(parameter.Name))
		writeKV(b, indent+4, "in", parameter.In)
		writeOptionalKV(b, indent+4, "description", parameter.Description)
		if parameter.Required {
			writeKV(b, indent+4, "required", "true")
		}
		writeIndent(b, indent+4)
		b.WriteString("schema:\n")
		writeSchema(b, indent+6, parameter.Schema)
	}
}

func writeResponses(b *bytes.Buffer, indent int, responses map[string]*openapi.Response) {
	writeIndent(b, indent)
	b.WriteString("responses:\n")
	for _, status := range sortedKeys(responses) {
		writeIndent(b, indent+2)
		fmt.Fprintf(b, "%s:\n", quoteString(status))
		response := responses[status]
		writeKV(b, indent+4, "description", response.Description)
		writeContent(b, indent+4, response.Content)
	}
}

func writeContent(b *bytes.Buffer, indent int, content map[string]*openapi.MediaType) {
	if len(content) == 0 {
		return
	}
	writeIndent(b, indent)
	b.WriteString("content:\n")
	for _, contentType := range sortedKeys(content) {
		writeIndent(b, indent+2)
		fmt.Fprintf(b, "%s:\n", quoteString(contentType))
		writeIndent(b, indent+4)
		b.WriteString("schema:\n")
		writeSchema(b, indent+6, content[contentType].Schema)
	}
}

func writeSchema(b *bytes.Buffer, indent int, schema *openapi.Schema) {
	writeOptionalKV(b, indent, "$ref", schema.Ref)
	if schema.Ref != "" {
		return
	}
	writeOptionalKV(b, indent, "type", schema.Type)
	writeOptionalKV(b, indent, "format", schema.Format)
	writeOptionalKV(b, indent, "description", schema.Description)
	if schema.Nullable {
		writeKV(b, indent, "nullable", "true")
	}
	if schema.Items != nil {
		writeIndent(b, indent)
		b.WriteString("items:\n")
		writeSchema(b, indent+2, schema.Items)
	}
	if len(schema.Properties) > 0 {
		writeIndent(b, indent)
		b.WriteString("properties:\n")
		for _, property := range sortedKeys(schema.Properties) {
			writeIndent(b, indent+2)
			fmt.Fprintf(b, "%s:\n", property)
			writeSchema(b, indent+4, schema.Properties[property])
		}
	}
	writeStringSlice(b, indent, "required", schema.Required)
	if schema.AdditionalProperties != nil {
		writeIndent(b, indent)
		b.WriteString("additionalProperties:\n")
		writeSchema(b, indent+2, schema.AdditionalProperties)
	}
}

func writeStringSlice(b *bytes.Buffer, indent int, key string, values []string) {
	if len(values) == 0 {
		return
	}
	writeIndent(b, indent)
	fmt.Fprintf(b, "%s:\n", key)
	for _, value := range values {
		writeIndent(b, indent+2)
		fmt.Fprintf(b, "- %s\n", quoteString(value))
	}
}

func writeOptionalKV(b *bytes.Buffer, indent int, key, value string) {
	if value == "" {
		return
	}
	writeKV(b, indent, key, value)
}

func writeKV(b *bytes.Buffer, indent int, key, value string) {
	writeIndent(b, indent)
	fmt.Fprintf(b, "%s: %s\n", key, quoteString(value))
}

func writeIndent(b *bytes.Buffer, n int) {
	b.WriteString(strings.Repeat(" ", n))
}

func quoteString(value string) string {
	if value == "" {
		return `""`
	}
	if value == "true" || value == "false" {
		return value
	}
	return fmt.Sprintf("%q", value)
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
