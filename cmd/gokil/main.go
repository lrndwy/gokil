package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gokil/config"
	"gokil/internal/scaffold"
	"gokil/migration"
	"gokil/orm"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "startproject":
		if err := startproject(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "serve":
		log.Fatal("serve must be run from your project: go run ./cmd/<project> serve")
	case "makemigrations":
		if err := makemigrations(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "migrate":
		if err := migrateCmd(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "doctor":
		if err := doctor(); err != nil {
			log.Fatal(err)
		}
	case "version":
		fmt.Println("gokil v0.1.0")
	default:
		usage()
		os.Exit(1)
	}
}

func startproject(args []string) error {
	flags := flag.NewFlagSet("startproject", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	dir := flags.String("dir", "", "output directory")

	// Reorder args so flags come before positional name (Go flag pkg stops at first non-flag).
	ordered := make([]string, 0, len(args))
	var name string
	for i := 0; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			ordered = append(ordered, args[i], args[i+1])
			i++
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			ordered = append(ordered, args[i])
			continue
		}
		if name == "" {
			name = args[i]
		}
	}
	_ = flags.Parse(ordered)

	if name == "" {
		name = flags.Arg(0)
	}
	if name == "" {
		return fmt.Errorf("usage: gokil startproject <name> [--dir path]")
	}

	outDir := *dir
	if outDir == "" {
		outDir = name
	}

	return scaffold.Create(scaffold.Options{
		Name:        name,
		Dir:         outDir,
		ModPath:     name,
		ReplacePath: computeReplacePath(outDir),
	})
}

func computeReplacePath(projectDir string) string {
	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		return ".."
	}
	absFramework, err := filepath.Abs(".")
	if err != nil {
		return ".."
	}
	rel, err := filepath.Rel(absProject, absFramework)
	if err != nil {
		return ".."
	}
	return rel
}

func makemigrations(args []string) error {
	flags := flag.NewFlagSet("makemigrations", flag.ContinueOnError)
	_ = flags.Parse(args)

	name := "auto"
	if flags.NArg() > 0 {
		name = flags.Arg(0)
	}

	settings, err := config.Load(config.Options{})
	if err != nil {
		return err
	}

	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required for makemigrations")
	}

	db, err := orm.Connect(
		settings.Database.Driver,
		settings.Database.DSN,
		settings.Database.MaxOpenConns,
		settings.Database.MaxIdleConns,
	)
	if err != nil {
		return err
	}
	defer db.Close()

	loadProjectModels()

	if len(orm.AllModels()) == 0 {
		return fmt.Errorf("no models registered — run from your project: go run ./cmd/<project> makemigrations")
	}

	detector := migration.Detector{DB: db.DB}
	diff, err := detector.Detect()
	if err != nil {
		return err
	}

	if !migration.HasChanges(diff) {
		fmt.Println("No changes detected")
		return nil
	}

	gen := migration.Generator{Dir: settings.Database.MigrationsDir}
	path, err := gen.GenerateFromDiff(diff, name)
	if err != nil {
		return err
	}

	fmt.Printf("Created migration: %s\n", path)
	return nil
}

func migrateCmd(args []string) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	rollback := flags.Bool("rollback", false, "rollback last migration")
	_ = flags.Parse(args)

	settings, err := config.Load(config.Options{})
	if err != nil {
		return err
	}

	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required for migrate")
	}

	db, err := orm.Connect(
		settings.Database.Driver,
		settings.Database.DSN,
		settings.Database.MaxOpenConns,
		settings.Database.MaxIdleConns,
	)
	if err != nil {
		return err
	}
	defer db.Close()

	runner := migration.Runner{
		DB:  db.DB,
		Dir: settings.Database.MigrationsDir,
	}

	if *rollback {
		if err := runner.Rollback(); err != nil {
			return err
		}
		fmt.Println("Rolled back last migration")
		return nil
	}

	count, err := runner.Migrate()
	if err != nil {
		return err
	}
	fmt.Printf("Applied %d migration(s)\n", count)
	return nil
}

func doctor() error {
	settings, err := config.Load(config.Options{})
	if err != nil {
		return err
	}

	if err := settings.Validate(); err != nil {
		return err
	}

	fmt.Println("Settings: OK")

	if settings.Database.DSN != "" {
		db, err := orm.Connect(
			settings.Database.Driver,
			settings.Database.DSN,
			settings.Database.MaxOpenConns,
			settings.Database.MaxIdleConns,
		)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		defer db.Close()
		fmt.Println("Database: OK")
	} else {
		fmt.Println("Database: skipped (GOKIL_DB_DSN not set)")
	}

	provider := settings.Storage.Provider
	if provider == "local" {
		path := settings.Storage.LocalPath
		if path == "" {
			path = "storage"
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("storage: %w", err)
		}
		fmt.Printf("Storage (local): OK (%s)\n", path)
	} else {
		fmt.Printf("Storage (%s): configured\n", provider)
	}

	return nil
}

func loadProjectModels() {
	// Import project models via blank import if models package exists.
	// When run from a project root, user should ensure models are registered via init().
	// Try loading ./_models_init.go or rely on project having imported models in main.
	_ = filepath.Clean(".")
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: gokil <command> [options]

Commands:
  startproject <name>   Create a new project
  makemigrations [name] Generate migration files from models
  migrate               Apply pending migrations
  migrate --rollback    Rollback last migration
  doctor                Validate configuration
  version               Show version`)
}
