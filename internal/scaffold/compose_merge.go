package scaffold

import (
	"fmt"
	"strings"
)

// MergeComposeServices merges "services:" blocks by simple text insertion.
//
// This is intentionally minimal (no YAML parser dependency). It supports:
// - If target has no "services:" key, it prepends add.
// - If target has "services:" but missing the service, it inserts it right after "services:" line.
// - If target already contains the service name under services, it returns target unchanged.
func MergeComposeServices(target string, serviceName string, add string) (string, error) {
	if strings.TrimSpace(add) == "" {
		return target, nil
	}
	if strings.TrimSpace(target) == "" {
		return add, nil
	}

	// quick check: already has service definition
	needle := "\n  " + serviceName + ":\n"
	if strings.Contains(target, needle) || strings.HasPrefix(target, "services:\n  "+serviceName+":\n") {
		return target, nil
	}

	lines := strings.Split(target, "\n")
	servicesIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "services:" {
			servicesIdx = i
			break
		}
	}

	if servicesIdx == -1 {
		// no services: -> just prepend add and keep target below
		return strings.TrimRight(add, "\n") + "\n\n" + strings.TrimLeft(target, "\n"), nil
	}

	// add should start with services:, but we only want the service content part.
	addLines := strings.Split(add, "\n")
	addServiceStart := -1
	for i, l := range addLines {
		if strings.TrimSpace(l) == "services:" {
			addServiceStart = i + 1
			break
		}
	}
	if addServiceStart == -1 {
		return "", fmt.Errorf("compose add block must contain services:")
	}
	serviceBlock := strings.Join(addLines[addServiceStart:], "\n")
	serviceBlock = strings.TrimLeft(serviceBlock, "\n")

	// insert right after services: line
	out := make([]string, 0, len(lines)+len(addLines))
	out = append(out, lines[:servicesIdx+1]...)
	out = append(out, strings.TrimRight(serviceBlock, "\n"))
	out = append(out, lines[servicesIdx+1:]...)

	merged := strings.Join(out, "\n")
	merged = strings.ReplaceAll(merged, "\n\n\n", "\n\n")
	return merged, nil
}

