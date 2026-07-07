package views

import (
	"sort"
	"strings"
)

// Required returns 400 Bad Request when value is empty.
func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return BadRequest(field + " is required")
	}
	return nil
}

// RequiredFields returns 400 Bad Request when one or more fields are empty.
func RequiredFields(fields map[string]string) error {
	missing := make([]string, 0, len(fields))
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return BadRequest("missing required fields: " + strings.Join(missing, ", "))
}
