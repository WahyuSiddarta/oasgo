package openapi

import (
	"strings"
	"testing"
)

func TestValidateAcceptsSupportedSubset(t *testing.T) {
	doc := NewDocument("Example API", "1.0.0")
	doc.Components.Schemas["UserResponse"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"id": {Type: "string"},
		},
		Required: []string{"id"},
	}
	doc.Paths["/users/{id}"] = &PathItem{Operations: map[string]*Operation{
		"get": {
			OperationID: "getUser",
			Parameters: []Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema:   &Schema{Type: "string"},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "User found.",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/UserResponse"},
						},
					},
				},
			},
		},
	}}

	if err := Validate(doc); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsMissingOperationResponses(t *testing.T) {
	doc := NewDocument("Example API", "1.0.0")
	doc.Paths["/users"] = &PathItem{Operations: map[string]*Operation{
		"get": {},
	}}

	err := Validate(doc)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "responses is required") {
		t.Fatalf("error = %q, want missing responses", err.Error())
	}
}

func TestValidateRejectsMissingPathParameterDeclaration(t *testing.T) {
	doc := NewDocument("Example API", "1.0.0")
	doc.Paths["/users/{id}"] = &PathItem{Operations: map[string]*Operation{
		"get": {
			Responses: map[string]*Response{
				"200": {Description: "OK."},
			},
		},
	}}

	err := Validate(doc)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), `missing required path parameter "id"`) {
		t.Fatalf("error = %q, want missing path parameter", err.Error())
	}
}

func TestValidateRejectsMissingSchemaReference(t *testing.T) {
	doc := NewDocument("Example API", "1.0.0")
	doc.Paths["/users"] = &PathItem{Operations: map[string]*Operation{
		"get": {
			Responses: map[string]*Response{
				"200": {
					Description: "OK.",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/Missing"},
						},
					},
				},
			},
		},
	}}

	err := Validate(doc)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), `missing schema "Missing"`) {
		t.Fatalf("error = %q, want missing schema reference", err.Error())
	}
}
