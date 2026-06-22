package render

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/wahyusiddarta/oasgo/internal/openapi"
)

// RenderYAML renders an OpenAPI document with deterministic field ordering.
func RenderYAML(doc *openapi.Document) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("oasgo: document is nil")
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
			if item == nil {
				continue
			}
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
	if operation == nil {
		return
	}
	writeOptionalKV(b, indent, "operationId", operation.OperationID)
	writeOptionalKV(b, indent, "summary", operation.Summary)
	writeOptionalKV(b, indent, "description", operation.Description)
	if len(operation.Tags) > 0 {
		writeIndent(b, indent)
		b.WriteString("tags:\n")
		for _, tag := range operation.Tags {
			writeIndent(b, indent+2)
			fmt.Fprintf(b, "- %s\n", quoteString(tag))
		}
	}
	if len(operation.Responses) > 0 {
		writeIndent(b, indent)
		b.WriteString("responses:\n")
		for _, status := range sortedKeys(operation.Responses) {
			writeIndent(b, indent+2)
			fmt.Fprintf(b, "%s:\n", quoteString(status))
			writeKV(b, indent+4, "description", operation.Responses[status].Description)
		}
	}
}

func writeSchema(b *bytes.Buffer, indent int, schema *openapi.Schema) {
	if schema == nil {
		writeKV(b, indent, "type", "object")
		return
	}
	writeOptionalKV(b, indent, "$ref", schema.Ref)
	writeOptionalKV(b, indent, "type", schema.Type)
	writeOptionalKV(b, indent, "format", schema.Format)
	writeOptionalKV(b, indent, "description", schema.Description)
	if schema.Nullable {
		writeKV(b, indent, "nullable", "true")
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
