package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	basePath string
	baseURL  string
}

func NewLocal(basePath, baseURL string) (*Local, error) {
	if strings.TrimSpace(basePath) == "" {
		basePath = "storage"
	}
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &Local{basePath: basePath, baseURL: baseURL}, nil
}

func (l *Local) Upload(ctx context.Context, key string, reader io.Reader, _ int64, _ string) error {
	path := l.filePath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

func (l *Local) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(l.filePath(key))
}

func (l *Local) Delete(ctx context.Context, key string) error {
	return os.Remove(l.filePath(key))
}

func (l *Local) URL(key string) (string, error) {
	if l.baseURL != "" {
		return strings.TrimSuffix(l.baseURL, "/") + "/" + key, nil
	}
	return "/storage/" + key, nil
}

func (l *Local) filePath(key string) string {
	return filepath.Join(l.basePath, filepath.Clean("/"+key))
}
