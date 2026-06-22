package operationcomment

import "testing"

func TestParseOperationBlock(t *testing.T) {
	op, err := Parse(Block{Lines: []string{
		"route: POST /users/{id}",
		"operationId: createUser",
		"summary: Create user",
		"tags:",
		"  - users",
		"parameters:",
		"  - name: id",
		"    in: path",
		"    type: string",
		"    required: true",
		"request:",
		"  body:",
		"    type: CreateUserRequest",
		"    required: true",
		"responses:",
		"  \"201\":",
		"    description: User created.",
		"    body:",
		"      type: UserResponse",
	}})
	if err != nil {
		t.Fatal(err)
	}

	if op.Route.Method != "POST" || op.Route.Path != "/users/{id}" {
		t.Fatalf("route = %#v", op.Route)
	}
	if op.OperationID != "createUser" {
		t.Fatalf("operationId = %q", op.OperationID)
	}
	if len(op.Tags) != 1 || op.Tags[0] != "users" {
		t.Fatalf("tags = %#v", op.Tags)
	}
	if len(op.Parameters) != 1 || op.Parameters[0].Name != "id" || !op.Parameters[0].Required {
		t.Fatalf("parameters = %#v", op.Parameters)
	}
	if op.Request == nil || op.Request.Body == nil || op.Request.Body.Type != "CreateUserRequest" {
		t.Fatalf("request = %#v", op.Request)
	}
	if op.Responses["201"].Description != "User created." {
		t.Fatalf("responses = %#v", op.Responses)
	}
	if op.Responses["201"].Body == nil || op.Responses["201"].Body.Type != "UserResponse" {
		t.Fatalf("response body = %#v", op.Responses["201"].Body)
	}
}

func TestParseRequiresRouteAndResponses(t *testing.T) {
	_, err := Parse(Block{Lines: []string{"summary: Missing route"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
