package operationcomment

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

	p := parser{lines: block.Lines}
	op, err := p.parse()
	if err != nil {
		return Operation{}, err
	}
	if op.Route.Method == "" || op.Route.Path == "" {
		return Operation{}, errors.New("oasgo: route is required")
	}
	if len(op.Responses) == 0 {
		return Operation{}, errors.New("oasgo: responses is required")
	}
	return op, nil
}

type parser struct {
	lines []string
	i     int
}

func (p *parser) parse() (Operation, error) {
	op := Operation{Responses: map[string]Response{}}
	for p.i < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.i])
		if line == "" {
			p.i++
			continue
		}

		key, value, ok := splitKeyValue(line)
		if !ok {
			return Operation{}, p.errf("expected key/value line")
		}

		switch key {
		case "route":
			route, err := parseRoute(value)
			if err != nil {
				return Operation{}, p.errf("%s", err.Error())
			}
			op.Route = route
			p.i++
		case "operationId":
			op.OperationID = value
			p.i++
		case "summary":
			op.Summary = value
			p.i++
		case "description":
			op.Description = value
			p.i++
		case "tags":
			tags, err := p.parseStringList(value, 2)
			if err != nil {
				return Operation{}, err
			}
			op.Tags = tags
		case "parameters":
			parameters, err := p.parseParameters()
			if err != nil {
				return Operation{}, err
			}
			op.Parameters = parameters
		case "request":
			request, err := p.parseRequest()
			if err != nil {
				return Operation{}, err
			}
			op.Request = request
		case "responses":
			responses, err := p.parseResponses()
			if err != nil {
				return Operation{}, err
			}
			op.Responses = responses
		default:
			return Operation{}, p.errf("unsupported field %q", key)
		}
	}
	return op, nil
}

func (p *parser) parseStringList(inline string, indent int) ([]string, error) {
	if inline != "" {
		values, err := parseInlineList(inline)
		if err != nil {
			return nil, p.errf("%s", err.Error())
		}
		p.i++
		return values, nil
	}

	p.i++
	var values []string
	for p.i < len(p.lines) {
		line := p.lines[p.i]
		if indentation(line) < indent || strings.TrimSpace(line) == "" {
			break
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, p.errf("expected list item")
		}
		values = append(values, cleanScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
		p.i++
	}
	return values, nil
}

func (p *parser) parseParameters() ([]Parameter, error) {
	p.i++
	var parameters []Parameter
	for p.i < len(p.lines) {
		if indentation(p.lines[p.i]) < 2 || strings.TrimSpace(p.lines[p.i]) == "" {
			break
		}
		trimmed := strings.TrimSpace(p.lines[p.i])
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, p.errf("expected parameter list item")
		}
		param := Parameter{}
		first := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if first != "" {
			if err := applyParameterField(&param, first); err != nil {
				return nil, p.errf("%s", err.Error())
			}
		}
		p.i++
		for p.i < len(p.lines) {
			if indentation(p.lines[p.i]) < 4 || strings.TrimSpace(p.lines[p.i]) == "" {
				break
			}
			if err := applyParameterField(&param, strings.TrimSpace(p.lines[p.i])); err != nil {
				return nil, p.errf("%s", err.Error())
			}
			p.i++
		}
		parameters = append(parameters, param)
	}
	return parameters, nil
}

func (p *parser) parseRequest() (*Request, error) {
	request := &Request{ContentType: "application/json"}
	p.i++
	for p.i < len(p.lines) {
		if indentation(p.lines[p.i]) < 2 || strings.TrimSpace(p.lines[p.i]) == "" {
			break
		}
		key, value, ok := splitKeyValue(strings.TrimSpace(p.lines[p.i]))
		if !ok {
			return nil, p.errf("expected request key/value line")
		}
		switch key {
		case "contentType":
			request.ContentType = value
			p.i++
		case "body":
			body, err := p.parseBody(4)
			if err != nil {
				return nil, err
			}
			request.Body = body
		default:
			return nil, p.errf("unsupported request field %q", key)
		}
	}
	return request, nil
}

func (p *parser) parseResponses() (map[string]Response, error) {
	responses := map[string]Response{}
	p.i++
	for p.i < len(p.lines) {
		if indentation(p.lines[p.i]) < 2 || strings.TrimSpace(p.lines[p.i]) == "" {
			break
		}
		status, _, ok := splitKeyValue(strings.TrimSpace(p.lines[p.i]))
		if !ok {
			return nil, p.errf("expected response status key")
		}
		response := Response{}
		p.i++
		for p.i < len(p.lines) {
			if indentation(p.lines[p.i]) < 4 || strings.TrimSpace(p.lines[p.i]) == "" {
				break
			}
			key, value, ok := splitKeyValue(strings.TrimSpace(p.lines[p.i]))
			if !ok {
				return nil, p.errf("expected response key/value line")
			}
			switch key {
			case "description":
				response.Description = value
				p.i++
			case "body":
				body, err := p.parseBody(6)
				if err != nil {
					return nil, err
				}
				response.Body = body
			default:
				return nil, p.errf("unsupported response field %q", key)
			}
		}
		responses[status] = response
	}
	return responses, nil
}

func (p *parser) parseBody(indent int) (*Body, error) {
	body := &Body{}
	p.i++
	for p.i < len(p.lines) {
		if indentation(p.lines[p.i]) < indent || strings.TrimSpace(p.lines[p.i]) == "" {
			break
		}
		key, value, ok := splitKeyValue(strings.TrimSpace(p.lines[p.i]))
		if !ok {
			return nil, p.errf("expected body key/value line")
		}
		switch key {
		case "type":
			body.Type = value
		case "required":
			required, err := strconv.ParseBool(value)
			if err != nil {
				return nil, p.errf("invalid body.required value %q", value)
			}
			body.Required = required
		case "description":
			body.Description = value
		default:
			return nil, p.errf("unsupported body field %q", key)
		}
		p.i++
	}
	return body, nil
}

func applyParameterField(param *Parameter, line string) error {
	key, value, ok := splitKeyValue(line)
	if !ok {
		return errors.New("expected parameter key/value line")
	}
	switch key {
	case "name":
		param.Name = value
	case "in":
		param.In = value
	case "type":
		param.Type = value
	case "required":
		required, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid parameter.required value %q", value)
		}
		param.Required = required
	case "description":
		param.Description = value
	default:
		return fmt.Errorf("unsupported parameter field %q", key)
	}
	return nil
}

func parseRoute(value string) (Route, error) {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return Route{}, errors.New("route must be METHOD /path")
	}
	return Route{Method: strings.ToUpper(fields[0]), Path: fields[1]}, nil
}

func splitKeyValue(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return cleanScalar(strings.TrimSpace(key)), cleanScalar(strings.TrimSpace(value)), true
}

func parseInlineList(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("expected inline list, got %q", value)
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if inner == "" {
		return nil, nil
	}
	parts := strings.Split(inner, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, cleanScalar(strings.TrimSpace(part)))
	}
	return values, nil
}

func cleanScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}

func indentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func (p *parser) errf(format string, args ...any) error {
	return fmt.Errorf("oasgo: line %d: %s", p.i+1, fmt.Sprintf(format, args...))
}
