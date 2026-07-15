package postman

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lrndwy/gokil/internal/routegen"
)

var (
	reRegister = regexp.MustCompile(`framework\.RegisterRoute\("([^"]+)",\s*"([^"]+)",\s*(\w+)\.(\w+)\)`)
	reLegacy   = regexp.MustCompile(`r\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)",\s*app\.Wrap\((\w+)\.(\w+)\)\)`)
	reJSONTag  = regexp.MustCompile(`json:"([^"]+)"`)
	reQueryStr = regexp.MustCompile(`ctx\.Query\("([^"]+)"\)`)
	reQueryInt = regexp.MustCompile(`ctx\.QueryInt\("([^"]+)"`)
)

// ParseProject scans the project directory and extracts route metadata.
func ParseProject(projectDir string) ([]RouteMetadata, error) {
	routes, err := parseRoutes(projectDir)
	if err != nil {
		return nil, err
	}

	handlers, err := parseAppHandlers(projectDir)
	if err != nil {
		// Fall back to legacy views/ if app/ handlers unavailable.
		handlers, _ = parseViews(projectDir)
	}

	for i := range routes {
		route := &routes[i]
		handlerName := route.Handler
		if idx := strings.LastIndex(handlerName, "."); idx >= 0 {
			handlerName = handlerName[idx+1:]
		}
		key := route.Method + " " + route.Path
		if h, ok := handlers[key]; ok {
			route.BodyFields = h.BodyFields
			route.QueryParams = h.QueryParams
		} else if h, ok := handlers[handlerName]; ok {
			route.BodyFields = h.BodyFields
			route.QueryParams = h.QueryParams
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routeOrder(routes[i].Method) < routeOrder(routes[j].Method)
	})

	return routes, nil
}

func routeOrder(method string) int {
	switch method {
	case "GET":
		return 0
	case "POST":
		return 1
	case "PUT":
		return 2
	case "PATCH":
		return 3
	case "DELETE":
		return 4
	default:
		return 5
	}
}

func parseRoutes(projectDir string) ([]RouteMetadata, error) {
	modPath, err := readModulePath(projectDir)
	if err == nil {
		appDir := filepath.Join(projectDir, "app")
		discovered, scanErr := routegen.Scan(appDir, modPath)
		if scanErr == nil && len(discovered) > 0 {
			routes := make([]RouteMetadata, 0, len(discovered))
			for _, r := range discovered {
				routes = append(routes, RouteMetadata{
					Method:  r.Method,
					Path:    r.Path,
					Handler: r.ImportAlias + "." + r.Handler,
				})
			}
			return routes, nil
		}
	}

	if routes, err := parseRegisterFile(projectDir); err == nil && len(routes) > 0 {
		return routes, nil
	}
	if routes, err := parseURLs(projectDir); err == nil && len(routes) > 0 {
		return routes, nil
	}
	return nil, fmt.Errorf("no routes found (expected app/**/route.go, app/register.go, or urls.go)")
}

func parseRegisterFile(projectDir string) ([]RouteMetadata, error) {
	content, err := readFirst(projectDir, filepath.Join("app", "register.go"))
	if err != nil {
		return nil, err
	}
	var routes []RouteMetadata
	for _, m := range reRegister.FindAllStringSubmatch(content, -1) {
		routes = append(routes, RouteMetadata{
			Method:  m[1],
			Path:    m[2],
			Handler: m[3] + "." + m[4],
		})
	}
	return routes, nil
}

// parseURLs reads urls.go and extracts registered routes (legacy).
func parseURLs(projectDir string) ([]RouteMetadata, error) {
	content, err := readFirst(projectDir, "urls.go")
	if err != nil {
		return nil, fmt.Errorf("cannot find urls.go: %w", err)
	}

	var routes []RouteMetadata
	matches := reLegacy.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		routes = append(routes, RouteMetadata{
			Method:  m[1],
			Path:    m[2],
			Handler: m[3] + "." + m[4],
		})
	}

	return routes, nil
}

func readFirst(projectDir string, name string) (string, error) {
	path := filepath.Join(projectDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readModulePath(projectDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found")
}

type handlerInfo struct {
	BodyFields  []Field
	QueryParams []QueryParam
}

// parseAppHandlers scans app/**/route.go for handler body/query metadata.
func parseAppHandlers(projectDir string) (map[string]handlerInfo, error) {
	appDir := filepath.Join(projectDir, "app")
	if _, err := os.Stat(appDir); err != nil {
		return nil, err
	}

	result := make(map[string]handlerInfo)
	err := filepath.WalkDir(appDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "route.go" {
			return nil
		}
		relDir, err := filepath.Rel(appDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		if relDir == "." {
			return nil
		}
		urlPath := folderToURL(relDir)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for funcName, funcBody := range extractFunctions(string(data)) {
			switch funcName {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
				result[funcName+" "+urlPath] = handlerInfo{
					BodyFields:  extractJSONFields(funcBody),
					QueryParams: extractQueryParams(funcBody),
				}
				result[funcName] = result[funcName+" "+urlPath]
			}
		}
		return nil
	})
	return result, err
}

func folderToURL(relDir string) string {
	parts := strings.Split(filepath.ToSlash(relDir), "/")
	segs := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		if strings.HasPrefix(p, "_") && len(p) > 1 {
			segs = append(segs, ":"+p[1:])
			continue
		}
		segs = append(segs, p)
	}
	if len(segs) == 0 {
		return "/"
	}
	return "/" + strings.Join(segs, "/")
}

// parseViews scans views/*.go files and extracts handler metadata (legacy).
func parseViews(projectDir string) (map[string]handlerInfo, error) {
	viewsDir := filepath.Join(projectDir, "views")
	entries, err := os.ReadDir(viewsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read views/: %w", err)
	}

	result := make(map[string]handlerInfo)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(viewsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		content := string(data)
		functions := extractFunctions(content)

		for funcName, funcBody := range functions {
			info := handlerInfo{}

			info.BodyFields = extractJSONFields(funcBody)
			info.QueryParams = extractQueryParams(funcBody)

			result[funcName] = info
		}
	}

	return result, nil
}

// extractFunctions splits Go source into individual function bodies.
func extractFunctions(content string) map[string]string {
	result := make(map[string]string)

	lines := strings.Split(content, "\n")
	var currentFunc string
	var braceCount int
	var bodyLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if currentFunc == "" {
			if strings.HasPrefix(trimmed, "func ") {
				// Extract function name: "func UserList(ctx *views.Context) error {"
				parts := strings.SplitN(trimmed, "(", 2)
				if len(parts) == 2 {
					nameParts := strings.Fields(parts[0])
					if len(nameParts) >= 2 {
						// Skip methods (func (r *Receiver) Name())
						if !strings.HasPrefix(nameParts[1], "(") {
							currentFunc = nameParts[1]
							bodyLines = nil
							braceCount = 0
						}
					}
				}
			}
		}

		if currentFunc != "" {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			bodyLines = append(bodyLines, line)

			if braceCount <= 0 && len(bodyLines) > 0 {
				result[currentFunc] = strings.Join(bodyLines, "\n")
				currentFunc = ""
				bodyLines = nil
			}
		}
	}

	return result
}

// extractJSONFields finds struct fields with json tags in a function body.
func extractJSONFields(funcBody string) []Field {
	lines := strings.Split(funcBody, "\n")
	collecting := false
	var structLines []string
	braceCount := 0

	// Join lines to detect multi-line patterns like "var input struct {"
	joined := strings.Join(lines, "\n")

	hasBodyStruct := strings.Contains(joined, "var input struct") ||
		strings.Contains(joined, "input := struct") ||
		strings.Contains(joined, "var body struct") ||
		strings.Contains(joined, "body := struct")
	if !hasBodyStruct {
		return nil
	}

	for _, line := range lines {
		if !collecting {
			if strings.Contains(line, "var input") || strings.Contains(line, "input :=") ||
				strings.Contains(line, "var body") || strings.Contains(line, "body :=") {
				collecting = true
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
				structLines = []string{line}
				if braceCount <= 0 {
					break
				}
			}
			continue
		}

		structLines = append(structLines, line)
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if braceCount <= 0 && len(structLines) > 1 {
			break
		}
	}

	if len(structLines) == 0 {
		return nil
	}

	var fields []Field
	for _, line := range structLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		matches := reJSONTag.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}

		for _, m := range matches {
			jsonName := m[1]
			if jsonName == "-" || jsonName == "" {
				continue
			}

			fieldType := inferFieldType(line)
			required := isRequiredField(jsonName, funcBody)

			fields = append(fields, Field{
				Name:     jsonName,
				Type:     fieldType,
				Required: required,
			})
		}
	}

	return fields
}

// inferFieldType guesses the Go type from a struct field line.
func inferFieldType(line string) string {
	trimmed := strings.TrimSpace(line)
	// Pattern: "Email string `json:\"email\"`"
	// or: "AuthorID int64 `json:\"author_id\"`"
	parts := strings.Fields(trimmed)
	if len(parts) >= 2 {
		goType := parts[1]
		switch {
		case strings.HasPrefix(goType, "int"):
			return "int64"
		case strings.HasPrefix(goType, "float"):
			return "float64"
		case goType == "bool":
			return "bool"
		default:
			return "string"
		}
	}
	return "string"
}

// isRequiredField checks if a field name appears in RequiredFields/Required calls.
func isRequiredField(fieldName, funcBody string) bool {
	// Look for RequiredFields(map[string]string{"field": ...})
	// or Required("field", ...)
	return strings.Contains(funcBody, `"`+fieldName+`"`) &&
		(strings.Contains(funcBody, "RequiredFields") || strings.Contains(funcBody, "Required("))
}

// extractQueryParams finds ctx.Query and ctx.QueryInt calls.
func extractQueryParams(funcBody string) []QueryParam {
	seen := make(map[string]bool)
	var params []QueryParam

	for _, m := range reQueryStr.FindAllStringSubmatch(funcBody, -1) {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			params = append(params, QueryParam{
				Key:         name,
				Value:       "",
				Description: "Query parameter",
			})
		}
	}

	for _, m := range reQueryInt.FindAllStringSubmatch(funcBody, -1) {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			params = append(params, QueryParam{
				Key:         name,
				Value:       "0",
				Description: "Query parameter (integer)",
			})
		}
	}

	return params
}
