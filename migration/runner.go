package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const migrationsTable = "gokil_migrations"

type Runner struct {
	DB  *sql.DB
	Dir string
}

type AppliedMigration struct {
	Name      string
	AppliedAt time.Time
}

func (r *Runner) EnsureMigrationsTable() error {
	_, err := r.DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`, migrationsTable))
	return err
}

func (r *Runner) Applied() (map[string]bool, error) {
	if err := r.EnsureMigrationsTable(); err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(fmt.Sprintf("SELECT name FROM %s ORDER BY name", migrationsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func (r *Runner) Pending() ([]string, error) {
	applied, err := r.Applied()
	if err != nil {
		return nil, err
	}

	files, err := listMigrationFiles(r.Dir)
	if err != nil {
		return nil, err
	}

	pending := []string{}
	for _, f := range files {
		if !applied[f] {
			pending = append(pending, f)
		}
	}
	return pending, nil
}

func (r *Runner) Migrate() (int, error) {
	pending, err := r.Pending()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, name := range pending {
		if err := r.apply(name); err != nil {
			return count, fmt.Errorf("apply %s: %w", name, err)
		}
		count++
	}
	return count, nil
}

func (r *Runner) Rollback() error {
	if err := r.EnsureMigrationsTable(); err != nil {
		return err
	}

	var lastName string
	err := r.DB.QueryRow(fmt.Sprintf(
		"SELECT name FROM %s ORDER BY applied_at DESC, name DESC LIMIT 1",
		migrationsTable,
	)).Scan(&lastName)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no migrations to rollback")
	}
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filepath.Join(r.Dir, lastName))
	if err != nil {
		return err
	}

	downSQL := extractSection(string(content), "Down")
	if strings.TrimSpace(downSQL) == "" {
		return fmt.Errorf("no down migration in %s", lastName)
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(downSQL); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE name = $1", migrationsTable), lastName); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Runner) apply(name string) error {
	content, err := os.ReadFile(filepath.Join(r.Dir, name))
	if err != nil {
		return err
	}

	upSQL := extractSection(string(content), "Up")
	if strings.TrimSpace(upSQL) == "" {
		return fmt.Errorf("no up migration in %s", name)
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(upSQL); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", migrationsTable), name); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	files := []string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	return files, nil
}

func ExtractSection(content, section string) string {
	return extractSection(content, section)
}

func extractSection(content, section string) string {
	marker := "-- +migrate " + section
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	start := idx + len(marker)
	rest := content[start:]

	// Find next section marker
	nextIdx := strings.Index(rest, "-- +migrate ")
	if nextIdx >= 0 {
		rest = rest[:nextIdx]
	}
	return strings.TrimSpace(rest)
}
