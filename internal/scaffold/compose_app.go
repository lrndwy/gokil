package scaffold

import (
	"bytes"
	"text/template"
)

type ComposeAppOptions struct {
	ServiceName string
}

const dockerComposeAppTemplate = `services:
  {{.Service}}:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        PROJECT: {{.Project}}
    env_file:
      - .env
    ports:
      - "${GOKIL_PORT:-8080}:8080"
{{- if .DependsOn }}
    depends_on:
{{- range .DependsOn }}
      - {{.}}
{{- end }}
{{- end }}
    restart: unless-stopped
`

// RenderDockerComposeApp renders a compose file containing only the gokil app service.
func RenderDockerComposeApp(projectName string, opts ComposeAppOptions, dependsOn []string) (string, error) {
	service := opts.ServiceName
	if service == "" {
		service = "gokil"
	}

	t, err := template.New("compose-app").Parse(dockerComposeAppTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]any{
		"Service":   service,
		"Project":   projectName,
		"DependsOn": dependsOn,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

