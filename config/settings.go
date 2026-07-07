package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Settings struct {
	AppName  string
	Env      string
	Debug    bool
	Host     string
	Port     int
	Database DatabaseSettings
	Redis    RedisSettings
	Storage  StorageSettings
}

type DatabaseSettings struct {
	Driver        string
	DSN           string
	Host          string
	Port          int
	User          string
	Password      string
	Name          string
	MaxOpenConns  int
	MaxIdleConns  int
	MigrationsDir string
}

type RedisSettings struct {
	Enabled  bool
	URL      string
	Host     string
	Port     int
	Password string
	DB       int
}

type StorageSettings struct {
	Provider        string
	LocalPath       string
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BaseURL         string
	UseSSL          bool
}

type Options struct {
	Prefix   string
	Lookuper func(string) (string, bool)
}

func Load(opts Options) (Settings, error) {
	lookup := opts.Lookuper
	if lookup == nil {
		loadDotEnv()
		lookup = os.LookupEnv
	}

	prefix := opts.Prefix
	if prefix == "" {
		prefix = "GOKIL"
	}

	settings := Settings{
		AppName: getString(lookup, prefix, "APP_NAME", "gokil"),
		Env:     getString(lookup, prefix, "ENV", "development"),
		Debug:   getBool(lookup, prefix, "DEBUG", true),
		Host:    getString(lookup, prefix, "HOST", "127.0.0.1"),
		Port:    getInt(lookup, prefix, "PORT", 8080),
		Database: DatabaseSettings{
			Driver:        getString(lookup, prefix, "DB_DRIVER", "postgres"),
			DSN:           getString(lookup, prefix, "DB_DSN", ""),
			Host:          getString(lookup, prefix, "DB_HOST", "localhost"),
			Port:          getInt(lookup, prefix, "DB_PORT", 5432),
			User:          getString(lookup, prefix, "DB_USER", ""),
			Password:      getString(lookup, prefix, "DB_PASSWORD", ""),
			Name:          getString(lookup, prefix, "DB_NAME", ""),
			MaxOpenConns:  getInt(lookup, prefix, "DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:  getInt(lookup, prefix, "DB_MAX_IDLE_CONNS", 5),
			MigrationsDir: getString(lookup, prefix, "DB_MIGRATIONS_DIR", "migrations"),
		},
		Redis: RedisSettings{
			Enabled:  getBool(lookup, prefix, "REDIS_ENABLED", false),
			URL:      getString(lookup, prefix, "REDIS_URL", ""),
			Host:     getString(lookup, prefix, "REDIS_HOST", "localhost"),
			Port:     getInt(lookup, prefix, "REDIS_PORT", 6379),
			Password: getString(lookup, prefix, "REDIS_PASSWORD", ""),
			DB:       getInt(lookup, prefix, "REDIS_DB", 0),
		},
		Storage: StorageSettings{
			Provider:        getString(lookup, prefix, "STORAGE_PROVIDER", "local"),
			LocalPath:       getString(lookup, prefix, "STORAGE_LOCAL_PATH", "storage"),
			Bucket:          getString(lookup, prefix, "STORAGE_BUCKET", ""),
			Region:          getString(lookup, prefix, "STORAGE_REGION", ""),
			Endpoint:        getString(lookup, prefix, "STORAGE_ENDPOINT", ""),
			AccessKeyID:     getString(lookup, prefix, "STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: getString(lookup, prefix, "STORAGE_SECRET_ACCESS_KEY", ""),
			BaseURL:         getString(lookup, prefix, "STORAGE_BASE_URL", ""),
			UseSSL:          getBool(lookup, prefix, "STORAGE_USE_SSL", true),
		},
	}

	if err := settings.Validate(); err != nil {
		return Settings{}, err
	}

	return settings, nil
}

func (s Settings) Validate() error {
	if strings.TrimSpace(s.AppName) == "" {
		return fmt.Errorf("app name is required")
	}
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(s.Storage.Provider) == "" {
		return fmt.Errorf("storage provider is required")
	}
	if strings.TrimSpace(s.Database.Driver) == "" {
		return fmt.Errorf("database driver is required")
	}
	return nil
}

func getString(lookup func(string) (string, bool), prefix, key, fallback string) string {
	if value, ok := lookup(envKey(prefix, key)); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func getInt(lookup func(string) (string, bool), prefix, key string, fallback int) int {
	value, ok := lookup(envKey(prefix, key))
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getBool(lookup func(string) (string, bool), prefix, key string, fallback bool) bool {
	value, ok := lookup(envKey(prefix, key))
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envKey(prefix, key string) string {
	return prefix + "_" + strings.ToUpper(key)
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
