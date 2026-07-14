package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Collection represents a Postman Collection v2.1.0.
type Collection struct {
	Info     Info       `json:"info"`
	Item     []Item     `json:"item"`
	Variable []Variable `json:"variable"`
}

// Info holds collection metadata.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// Item is either a folder (with nested Items) or a request.
type Item struct {
	Name     string     `json:"name"`
	Item     []Item     `json:"item,omitempty"`
	Request  *Request   `json:"request,omitempty"`
	Response *Response  `json:"response,omitempty"`
}

// Request defines an HTTP request.
type Request struct {
	Method  string    `json:"method"`
	Header  []Header  `json:"header"`
	Body    *Body     `json:"body,omitempty"`
	URL     URL       `json:"url"`
	Description string `json:"description,omitempty"`
}

// Header is an HTTP header.
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Body represents the request body.
type Body struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

// URL holds URL components.
type URL struct {
	Raw       string          `json:"raw"`
	Host      []string        `json:"host"`
	Path      []string        `json:"path"`
	Variable  []PathVariable  `json:"variable,omitempty"`
	Query     []QueryParam    `json:"query,omitempty"`
}

// PathVariable is a URL path parameter.
type PathVariable struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// QueryParam is a URL query parameter.
type QueryParam struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// Variable is a collection-level variable.
type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Response is a saved example response.
type Response struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Header []Header `json:"header"`
	Body   string `json:"body"`
}

// RouteMetadata describes a single API endpoint.
type RouteMetadata struct {
	Method      string
	Path        string
	Handler     string
	Description string
	Headers     []Header
	BodyFields  []Field
	PathParams  []PathParam
	QueryParams []QueryParam
}

// Field describes a JSON body field.
type Field struct {
	Name     string
	Type     string
	Required bool
}

// PathParam describes a URL path parameter.
type PathParam struct {
	Name string
	Type string
}

// Generate creates a Postman Collection from route metadata.
func Generate(projectName string, routes []RouteMetadata, baseURL string) *Collection {
	collection := &Collection{
		Info: Info{
			Name:        projectName,
			Description: fmt.Sprintf("Auto-generated Postman collection for %s", projectName),
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Variable: []Variable{
			{Key: "base_url", Value: baseURL},
			{Key: "token", Value: "your-jwt-token"},
		},
	}

	folders := groupByFolder(routes)
	folderNames := make([]string, 0, len(folders))
	for name := range folders {
		folderNames = append(folderNames, name)
	}
	sort.Strings(folderNames)

	for _, folderName := range folderNames {
		items := folders[folderName]
		folder := Item{
			Name: folderName,
			Item: make([]Item, 0, len(items)),
		}

		for _, route := range items {
			folder.Item = append(folder.Item, buildItem(route))
		}

		collection.Item = append(collection.Item, folder)
	}

	return collection
}

// Write writes the collection to a JSON file.
func Write(path string, collection *Collection) error {
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal collection: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// groupByFolder groups routes by the first path segment after /api/.
func groupByFolder(routes []RouteMetadata) map[string][]RouteMetadata {
	folders := make(map[string][]RouteMetadata)

	for _, route := range routes {
		folder := inferFolder(route)
		folders[folder] = append(folders[folder], route)
	}

	return folders
}

// inferFolder determines the folder name from the route path.
func inferFolder(route RouteMetadata) string {
	path := strings.TrimPrefix(route.Path, "/")
	parts := strings.Split(path, "/")

	// /api/users/ → "Users"
	// /api/users/:id → "Users"
	// /api/health/ → "Health"
	// /healthz → "General"
	if len(parts) >= 2 && parts[0] == "api" {
		return titleCase(parts[1])
	}
	if len(parts) >= 1 {
		return titleCase(parts[0])
	}
	return "General"
}

// titleCase capitalizes the first letter.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// buildItem creates a Postman Item from RouteMetadata.
func buildItem(route RouteMetadata) Item {
	name := inferName(route)
	req := buildRequest(route)

	return Item{
		Name:    name,
		Request: &req,
	}
}

// inferName creates a human-readable name for the route.
func inferName(route RouteMetadata) string {
	path := strings.TrimPrefix(route.Path, "/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	// Skip "api" prefix
	start := 0
	if len(parts) > 0 && parts[0] == "api" {
		start = 1
	}

	var segments []string
	for _, p := range parts[start:] {
		if strings.HasPrefix(p, ":") {
			segments = append(segments, "{id}")
		} else {
			segments = append(segments, titleCase(p))
		}
	}

	resource := strings.Join(segments, " ")

	action := actionForMethod(route.Method)

	// For GET with path param, use "Get" action
	if route.Method == "GET" && strings.Contains(route.Path, ":") {
		action = "Get"
	}

	if resource == "" {
		return fmt.Sprintf("%s /%s", route.Method, strings.Join(parts, "/"))
	}

	if action != "" {
		return fmt.Sprintf("%s %s", action, resource)
	}
	return resource
}

// actionForMethod returns an action verb for the HTTP method.
func actionForMethod(method string) string {
	switch method {
	case "GET":
		return ""
	case "POST":
		return "Create"
	case "PUT":
		return "Update"
	case "PATCH":
		return "Patch"
	case "DELETE":
		return "Delete"
	default:
		return method
	}
}

// buildRequest creates a Postman Request from RouteMetadata.
func buildRequest(route RouteMetadata) Request {
	headers := defaultHeaders()
	if len(route.Headers) > 0 {
		headers = append(headers, route.Headers...)
	}

	req := Request{
		Method:  route.Method,
		Header:  headers,
		URL:     buildURL(route),
	}

	if route.Description != "" {
		req.Description = route.Description
	}

	if needsBody(route.Method) && len(route.BodyFields) > 0 {
		raw := buildBodyJSON(route.BodyFields)
		req.Body = &Body{
			Mode: "raw",
			Raw:  raw,
		}
	}

	return req
}

// defaultHeaders returns the standard headers.
func defaultHeaders() []Header {
	return []Header{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Authorization", Value: "Bearer {{token}}"},
	}
}

// needsBody returns true if the method typically has a request body.
func needsBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// buildURL constructs the URL object.
func buildURL(route RouteMetadata) URL {
	host := []string{"{{base_url}}"}
	rawPath := strings.TrimPrefix(route.Path, "/")
	rawPath = strings.TrimSuffix(rawPath, "/")
	pathParts := strings.Split(rawPath, "/")

	var variables []PathVariable
	var queryParams []QueryParam

	for _, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			key := strings.TrimPrefix(part, ":")
			variables = append(variables, PathVariable{
				Key:         key,
				Value:       fmt.Sprintf(":%s", key),
				Description: fmt.Sprintf("Path parameter: %s", key),
			})
		}
	}

	for _, qp := range route.QueryParams {
		queryParams = append(queryParams, qp)
	}

	raw := "{{base_url}}/" + rawPath

	return URL{
		Raw:      raw,
		Host:     host,
		Path:     pathParts,
		Variable: variables,
		Query:    queryParams,
	}
}

// buildBodyJSON creates a JSON string from body fields.
func buildBodyJSON(fields []Field) string {
	obj := make(map[string]any)
	for _, f := range fields {
		obj[f.Name] = exampleValue(f.Type, f.Name)
	}

	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// exampleValue returns a placeholder value based on the type.
func exampleValue(typ, name string) any {
	switch typ {
	case "int", "int64":
		return 1
	case "bool":
		return true
	case "float64":
		return 0.0
	default:
		if strings.Contains(name, "email") {
			return "user@example.com"
		}
		if strings.Contains(name, "id") {
			return 1
		}
		return "string"
	}
}
