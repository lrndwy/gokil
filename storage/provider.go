package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"gokil/config"
)

type Provider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(key string) (string, error)
}

func New(settings config.StorageSettings) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(settings.Provider)) {
	case "", "local":
		return NewLocal(settings.LocalPath, settings.BaseURL)
	case "s3":
		return NewS3(settings)
	default:
		return nil, fmt.Errorf("unknown storage provider: %s", settings.Provider)
	}
}
