package scaffold

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	DatabasePostgres = "postgres"
	DatabaseMySQL    = "mysql"
)

var SupportedDatabases = []string{DatabasePostgres, DatabaseMySQL}

// InfraOptions controls optional Docker infrastructure generated for a project.
type InfraOptions struct {
	SetupDatabase bool
	Database      string
	SetupRedis    bool
}

// TemplateData is passed to project file templates.
type TemplateData struct {
	Name             string
	ModPath          string
	ReplacePath      string
	UseLocalReplace  bool
	FrameworkVersion string
	Infra            InfraConfig
}

// InfraConfig holds rendered infrastructure values for templates.
type InfraConfig struct {
	SetupDatabase bool
	SetupRedis    bool
	Database      string
	DBDriver      string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBDSN         string
	RedisEnabled  bool
	RedisHost     string
	RedisPort     int
	RedisURL      string
}

func DefaultInfraConfig(projectName string) InfraConfig {
	return BuildInfraConfig(projectName, InfraOptions{
		SetupDatabase: false,
		Database:      DatabasePostgres,
		SetupRedis:      false,
	})
}

func BuildInfraConfig(projectName string, opts InfraOptions) InfraConfig {
	db := normalizeDatabase(opts.Database)
	cfg := InfraConfig{
		SetupDatabase: opts.SetupDatabase,
		SetupRedis:    opts.SetupRedis,
		Database:      db,
		DBDriver:      db,
		DBHost:        "localhost",
		DBPort:        defaultDBPort(db),
		DBUser:        projectName,
		DBPassword:    "gokil",
		DBName:        projectName,
		RedisEnabled:  opts.SetupRedis,
		RedisHost:     "localhost",
		RedisPort:     6379,
	}
	cfg.DBDSN = BuildDSN(cfg)
	cfg.RedisURL = BuildRedisURL(cfg)
	return cfg
}

func normalizeDatabase(value string) string {
	switch value {
	case DatabaseMySQL, "mariadb":
		return DatabaseMySQL
	default:
		return DatabasePostgres
	}
}

func defaultDBPort(database string) int {
	if database == DatabaseMySQL {
		return 3306
	}
	return 5432
}

func BuildDSN(cfg InfraConfig) string {
	switch cfg.DBDriver {
	case DatabaseMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	default:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			url.QueryEscape(cfg.DBUser),
			url.QueryEscape(cfg.DBPassword),
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	}
}

func BuildRedisURL(cfg InfraConfig) string {
	return "redis://" + cfg.RedisHost + ":" + strconv.Itoa(cfg.RedisPort) + "/0"
}

func (c InfraConfig) NeedsDockerCompose() bool {
	return c.SetupDatabase || c.SetupRedis
}
