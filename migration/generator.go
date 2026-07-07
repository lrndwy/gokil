package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lrndwy/gokil/orm"
)

type Generator struct {
	Dir   string
	Clock func() time.Time
}

func (g Generator) GenerateFromDiff(diff *SchemaDiff, name string) (string, error) {
	if strings.TrimSpace(g.Dir) == "" {
		g.Dir = "migrations"
	}
	if g.Clock == nil {
		g.Clock = time.Now
	}
	if strings.TrimSpace(name) == "" {
		name = "auto"
	}

	if err := os.MkdirAll(g.Dir, 0o755); err != nil {
		return "", err
	}

	up, down := RenderDiff(diff)
	content := up + "\n" + down

	fileName := fmt.Sprintf("%s_%s.sql", g.Clock().UTC().Format("20060102150405"), slug(name))
	path := filepath.Join(g.Dir, fileName)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (g Generator) GenerateInitial(meta orm.ModelMeta) (string, error) {
	diff := &SchemaDiff{CreateTables: []orm.ModelMeta{meta}}
	return g.GenerateFromDiff(diff, slug(meta.Name))
}

func slug(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}
