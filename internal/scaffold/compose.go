package scaffold

import (
	"bytes"
	"fmt"
	"text/template"
)

const dockerComposeTemplate = `services:{{if .Infra.SetupDatabase}}
  db:
    image: {{if eq .Infra.DBDriver "mysql"}}mysql:8.4{{else}}postgres:16-alpine{{end}}
    restart: unless-stopped
    ports:
      - "${GOKIL_DB_PORT:-{{.Infra.DBPort}}}:{{.Infra.DBPort}}"
    environment:{{if eq .Infra.DBDriver "mysql"}}
      MYSQL_ROOT_PASSWORD: ${GOKIL_DB_PASSWORD:-{{.Infra.DBPassword}}}
      MYSQL_DATABASE: ${GOKIL_DB_NAME:-{{.Infra.DBName}}}
      MYSQL_USER: ${GOKIL_DB_USER:-{{.Infra.DBUser}}}
      MYSQL_PASSWORD: ${GOKIL_DB_PASSWORD:-{{.Infra.DBPassword}}}{{else}}
      POSTGRES_USER: ${GOKIL_DB_USER:-{{.Infra.DBUser}}}
      POSTGRES_PASSWORD: ${GOKIL_DB_PASSWORD:-{{.Infra.DBPassword}}}
      POSTGRES_DB: ${GOKIL_DB_NAME:-{{.Infra.DBName}}}{{end}}
    volumes:
      - db_data:{{if eq .Infra.DBDriver "mysql"}}/var/lib/mysql{{else}}/var/lib/postgresql/data{{end}}
    healthcheck:{{if eq .Infra.DBDriver "mysql"}}
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1"]
      interval: 5s
      timeout: 5s
      retries: 10{{else}}
      test: ["CMD-SHELL", "pg_isready -U ${GOKIL_DB_USER:-{{.Infra.DBUser}}} -d ${GOKIL_DB_NAME:-{{.Infra.DBName}}}"]
      interval: 5s
      timeout: 5s
      retries: 10{{end}}{{end}}{{if .Infra.SetupRedis}}
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "${GOKIL_REDIS_PORT:-{{.Infra.RedisPort}}}:6379"
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 10{{end}}

volumes:{{if .Infra.SetupDatabase}}
  db_data:{{end}}{{if .Infra.SetupRedis}}
  redis_data:{{end}}
`

func RenderDockerCompose(data TemplateData) (string, error) {
	t, err := template.New("compose").Parse(dockerComposeTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func validateInfra(data TemplateData) error {
	if data.Infra.SetupDatabase && data.Infra.DBDriver != DatabasePostgres && data.Infra.DBDriver != DatabaseMySQL {
		return fmt.Errorf("unsupported database: %s", data.Infra.DBDriver)
	}
	return nil
}
