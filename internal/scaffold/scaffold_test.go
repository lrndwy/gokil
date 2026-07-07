package scaffold_test

import (
	"strings"
	"testing"

	"github.com/lrndwy/gokil/internal/scaffold"
)

func TestBuildDSNPostgres(t *testing.T) {
	cfg := scaffold.BuildInfraConfig("myapi", scaffold.InfraOptions{
		SetupDatabase: true,
		Database:      scaffold.DatabasePostgres,
	})
	want := "postgres://myapi:gokil@localhost:5432/myapi?sslmode=disable"
	if cfg.DBDSN != want {
		t.Fatalf("DSN = %q, want %q", cfg.DBDSN, want)
	}
}

func TestBuildDSNMySQL(t *testing.T) {
	cfg := scaffold.BuildInfraConfig("myapi", scaffold.InfraOptions{
		SetupDatabase: true,
		Database:      scaffold.DatabaseMySQL,
	})
	want := "myapi:gokil@tcp(localhost:3306)/myapi?parseTime=true&charset=utf8mb4"
	if cfg.DBDSN != want {
		t.Fatalf("DSN = %q, want %q", cfg.DBDSN, want)
	}
}

func TestRenderDockerComposePostgresAndRedis(t *testing.T) {
	data := scaffold.TemplateData{
		Name: "myapi",
		Infra: scaffold.BuildInfraConfig("myapi", scaffold.InfraOptions{
			SetupDatabase: true,
			Database:      scaffold.DatabasePostgres,
			SetupRedis:    true,
		}),
	}
	compose, err := scaffold.RenderDockerCompose(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"postgres:16-alpine",
		"${GOKIL_DB_USER:-myapi}",
		"redis:7-alpine",
		"${GOKIL_REDIS_PORT:-6379}",
		"db_data:",
		"redis_data:",
	} {
		if !strings.Contains(compose, want) {
			t.Fatalf("compose missing %q:\n%s", want, compose)
		}
	}
}

func TestRenderDockerComposeMySQL(t *testing.T) {
	data := scaffold.TemplateData{
		Name: "shop",
		Infra: scaffold.BuildInfraConfig("shop", scaffold.InfraOptions{
			SetupDatabase: true,
			Database:      scaffold.DatabaseMySQL,
		}),
	}
	compose, err := scaffold.RenderDockerCompose(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(compose, "mysql:8.4") {
		t.Fatalf("expected mysql image:\n%s", compose)
	}
	if strings.Contains(compose, "redis:") {
		t.Fatal("redis service should not be generated")
	}
}
