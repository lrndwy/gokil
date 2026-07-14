package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lrndwy/gokil/cliui"
	"github.com/lrndwy/gokil/config"
	"github.com/lrndwy/gokil/internal/scaffold"
	"github.com/lrndwy/gokil/migration"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/postman"

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
	case "compose":
		if err := composeCmd(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "build":
		if err := buildCmd(os.Args[2:]); err != nil {
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
	case "postman":
		if err := postmanCmd(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "doctor":
		if err := doctor(); err != nil {
			log.Fatal(err)
		}
	case "version":
		printVersion()
	default:
		usage()
		os.Exit(1)
	}
}

func startproject(args []string) error {
	flags := flag.NewFlagSet("startproject", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	dir := flags.String("dir", "", "output directory")
	withDB := flags.Bool("db", false, "setup database with Docker Compose")
	noDB := flags.Bool("no-db", false, "skip database Docker Compose setup")
	dbEngine := flags.String("db-engine", "postgres", "database engine: postgres or mysql")
	withRedis := flags.Bool("redis", false, "setup Redis with Docker Compose")
	noRedis := flags.Bool("no-redis", false, "skip Redis Docker Compose setup")

	// Reorder args so flags come before positional name (Go flag pkg stops at first non-flag).
	ordered := make([]string, 0, len(args))
	var name string
	for i := 0; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			ordered = append(ordered, args[i], args[i+1])
			i++
			continue
		}
		if args[i] == "--db-engine" && i+1 < len(args) {
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
		return fmt.Errorf("usage: gokil startproject <name> [--dir path] [--db|--no-db] [--db-engine postgres|mysql] [--redis|--no-redis]")
	}
	if *withDB && *noDB {
		return fmt.Errorf("use only one of --db or --no-db")
	}
	if *withRedis && *noRedis {
		return fmt.Errorf("use only one of --redis or --no-redis")
	}

	outDir := *dir
	if outDir == "" {
		outDir = name
	}

	var infraPreset *scaffold.InfraOptions
	if *withDB || *noDB || *withRedis || *noRedis {
		infraPreset = &scaffold.InfraOptions{
			SetupDatabase: *withDB,
			Database:      *dbEngine,
			SetupRedis:    *withRedis,
		}
		if *noDB {
			infraPreset.SetupDatabase = false
		}
		if *noRedis {
			infraPreset.SetupRedis = false
		}
	}

	fmt.Println(cliui.Cyan("gokil") + " " + cliui.Dim("startproject") + " " + cliui.Bold(name))
	fmt.Println()

	return scaffold.Create(scaffold.Options{
		Name:    name,
		Dir:     outDir,
		ModPath: name,
		Infra:   infraPreset,
	})
}

func makemigrations(args []string) error {
	flags := flag.NewFlagSet("makemigrations", flag.ContinueOnError)
	_ = flags.Parse(args)

	name := "auto"
	if flags.NArg() > 0 {
		name = flags.Arg(0)
	}

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading configuration")

	settings, err := config.Load(config.Options{})
	if err != nil {
		sp.Fail("Loading configuration")
		return err
	}
	sp.Success("Loaded configuration")

	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required for makemigrations")
	}

	sp.Start("Connecting to database")
	db, err := orm.Connect(
		settings.Database.Driver,
		settings.Database.DSN,
		settings.Database.MaxOpenConns,
		settings.Database.MaxIdleConns,
	)
	if err != nil {
		sp.Fail("Connecting to database")
		return err
	}
	defer db.Close()
	sp.Success("Connected to database")

	loadProjectModels()

	if len(orm.AllModels()) == 0 {
		return fmt.Errorf("no models registered — run from your project: go run ./cmd/<project> makemigrations")
	}

	sp.Start("Detecting schema changes")
	detector := migration.Detector{DB: db.DB}
	diff, err := detector.Detect()
	if err != nil {
		sp.Fail("Detecting schema changes")
		return err
	}
	sp.Success("Detected schema changes")

	if !migration.HasChanges(diff) {
		cliui.Infof("No changes detected")
		return nil
	}

	sp.Start("Generating migration files")
	gen := migration.Generator{Dir: settings.Database.MigrationsDir}
	path, err := gen.GenerateFromDiff(diff, name)
	if err != nil {
		sp.Fail("Generating migration files")
		return err
	}
	sp.Success(fmt.Sprintf("Created migration: %s", path))
	return nil
}

func migrateCmd(args []string) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	rollback := flags.Bool("rollback", false, "rollback last migration")
	_ = flags.Parse(args)

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading configuration")

	settings, err := config.Load(config.Options{})
	if err != nil {
		sp.Fail("Loading configuration")
		return err
	}
	sp.Success("Loaded configuration")

	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required for migrate")
	}

	sp.Start("Connecting to database")
	db, err := orm.Connect(
		settings.Database.Driver,
		settings.Database.DSN,
		settings.Database.MaxOpenConns,
		settings.Database.MaxIdleConns,
	)
	if err != nil {
		sp.Fail("Connecting to database")
		return err
	}
	defer db.Close()
	sp.Success("Connected to database")

	runner := migration.Runner{
		DB:  db.DB,
		Dir: settings.Database.MigrationsDir,
	}

	if *rollback {
		sp.Start("Rolling back last migration")
		if err := runner.Rollback(); err != nil {
			sp.Fail("Rolling back last migration")
			return err
		}
		sp.Success("Rolled back last migration")
		return nil
	}

	sp.Start("Applying migrations")
	count, err := runner.Migrate()
	if err != nil {
		sp.Fail("Applying migrations")
		return err
	}
	if count == 0 {
		sp.Success("No pending migrations")
		return nil
	}
	sp.Success(fmt.Sprintf("Applied %d migration(s)", count))
	return nil
}

func doctor() error {
	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading configuration")

	settings, err := config.Load(config.Options{})
	if err != nil {
		sp.Fail("Loading configuration")
		return err
	}

	if err := settings.Validate(); err != nil {
		sp.Fail("Validating settings")
		return err
	}
	sp.Success("Settings validated")

	if settings.Database.DSN != "" {
		sp.Start("Checking database connection")
		db, err := orm.Connect(
			settings.Database.Driver,
			settings.Database.DSN,
			settings.Database.MaxOpenConns,
			settings.Database.MaxIdleConns,
		)
		if err != nil {
			sp.Fail("Checking database connection")
			return fmt.Errorf("database: %w", err)
		}
		defer db.Close()
		sp.Success("Database connection OK")
	} else {
		cliui.Warnf("Database skipped (GOKIL_DB_DSN not set)")
	}

	sp.Start("Checking storage")
	provider := settings.Storage.Provider
	if provider == "local" {
		path := settings.Storage.LocalPath
		if path == "" {
			path = "storage"
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			sp.Fail("Checking storage")
			return fmt.Errorf("storage: %w", err)
		}
		sp.Success(fmt.Sprintf("Storage ready (%s)", path))
	} else {
		sp.Success(fmt.Sprintf("Storage configured (%s)", provider))
	}

	return nil
}

func loadProjectModels() {
	// Import project models via blank import if models package exists.
	// When run from a project root, user should ensure models are registered via init().
	// Try loading ./_models_init.go or rely on project having imported models in main.
	_ = filepath.Clean(".")
}

func detectProjectName(projectFlag string) (string, error) {
	if strings.TrimSpace(projectFlag) != "" {
		return strings.TrimSpace(projectFlag), nil
	}

	entries, err := os.ReadDir("cmd")
	if err != nil {
		return "", fmt.Errorf("cannot detect project: missing cmd/ directory (run from project root or pass --project)")
	}

	var candidates []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		mainPath := filepath.Join("cmd", name, "main.go")
		if _, err := os.Stat(mainPath); err == nil {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("cannot detect project: no cmd/<project>/main.go found (pass --project)")
	}
	return "", fmt.Errorf("cannot detect project: multiple cmd/<project> found (%s) — pass --project", strings.Join(candidates, ", "))
}

func composeCmd(args []string) error {
	flags := flag.NewFlagSet("compose", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	project := flags.String("project", "", "project name (cmd/<project>)")
	out := flags.String("out", "docker-compose.yml", "output compose path")
	service := flags.String("service", "gokil", "app service name")
	update := flags.Bool("update", true, "update existing compose if present")
	onlyApp := flags.Bool("only-app", false, "generate compose containing only the app service")
	_ = flags.Parse(args)

	p, err := detectProjectName(*project)
	if err != nil {
		return err
	}

	// Ensure Dockerfile exists (generate if missing).
	if _, err := os.Stat("Dockerfile"); err != nil {
		df, err := scaffold.RenderDockerfile(scaffold.TemplateData{Name: p})
		if err != nil {
			return err
		}
		if err := os.WriteFile("Dockerfile", []byte(df), 0o644); err != nil {
			return err
		}
		cliui.Successf("Created Dockerfile")
	}

	dependsOn := make([]string, 0, 2)
	// If user already has infra compose, we'll auto-depend on db/redis if present.
	if existing, err := os.ReadFile(*out); err == nil {
		s := string(existing)
		if strings.Contains(s, "\n  db:\n") || strings.HasPrefix(s, "services:\n  db:\n") {
			dependsOn = append(dependsOn, "db")
		}
		if strings.Contains(s, "\n  redis:\n") || strings.HasPrefix(s, "services:\n  redis:\n") {
			dependsOn = append(dependsOn, "redis")
		}
	}

	appCompose, err := scaffold.RenderDockerComposeApp(p, scaffold.ComposeAppOptions{ServiceName: *service}, dependsOn)
	if err != nil {
		return err
	}

	if *onlyApp {
		if err := os.WriteFile(*out, []byte(appCompose), 0o644); err != nil {
			return err
		}
		cliui.Successf("Wrote %s", *out)
		return nil
	}

	// Update or create compose.
	if existing, err := os.ReadFile(*out); err == nil {
		if !*update {
			return fmt.Errorf("%s already exists (use --update or choose another --out)", *out)
		}
		merged, err := scaffold.MergeComposeServices(string(existing), *service, appCompose)
		if err != nil {
			return err
		}
		if err := os.WriteFile(*out, []byte(merged), 0o644); err != nil {
			return err
		}
		cliui.Successf("Updated %s (added service %q)", *out, *service)
		return nil
	}

	if err := os.WriteFile(*out, []byte(appCompose), 0o644); err != nil {
		return err
	}
	cliui.Successf("Created %s", *out)
	return nil
}

func buildCmd(args []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	project := flags.String("project", "", "project name (cmd/<project>)")
	out := flags.String("o", "", "output binary path (default: ./bin/<project>)")
	goos := flags.String("os", "", "GOOS (optional)")
	goarch := flags.String("arch", "", "GOARCH (optional)")
	_ = flags.Parse(args)

	p, err := detectProjectName(*project)
	if err != nil {
		return err
	}

	output := *out
	if strings.TrimSpace(output) == "" {
		output = filepath.Join("bin", p)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", output, "./cmd/"+p)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if *goos != "" {
		cmd.Env = append(cmd.Env, "GOOS="+*goos)
	}
	if *goarch != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+*goarch)
	}

	cliui.Infof("Building %s -> %s", p, output)
	if err := cmd.Run(); err != nil {
		return err
	}
	cliui.Successf("Build OK: %s", output)
	return nil
}

func postmanCmd(args []string) error {
	flags := flag.NewFlagSet("postman", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	project := flags.String("project", "", "project name (cmd/<project>)")
	out := flags.String("output", "collection_postman.json", "output file path")
	baseURL := flags.String("base-url", "http://localhost:8080", "base URL for API requests")
	_ = flags.Parse(args)

	p, err := detectProjectName(*project)
	if err != nil {
		return err
	}

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Parsing source files")

	routes, err := postman.ParseProject(".")
	if err != nil {
		sp.Fail("Parsing source files")
		return err
	}
	sp.Success(fmt.Sprintf("Found %d endpoint(s)", len(routes)))

	if len(routes) == 0 {
		cliui.Warnf("No routes found in urls.go")
		return nil
	}

	sp.Start("Generating Postman collection")
	collection := postman.Generate(p, routes, *baseURL)
	sp.Success("Generated Postman collection")

	sp.Start("Writing output file")
	if err := postman.Write(*out, collection); err != nil {
		sp.Fail("Writing output file")
		return err
	}
	sp.Success(fmt.Sprintf("Written to %s", *out))

	cliui.Infof("Import the collection into Postman to use it")
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, cliui.Bold("Usage: gokil <command> [options]"))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, cliui.Cyan("Commands:"))
	fmt.Fprintln(os.Stderr, `  startproject <name>   Create a new project
                          --db / --no-db
                          --db-engine postgres|mysql
                          --redis / --no-redis
  compose              Generate/update docker-compose.yml with gokil service
  build                Compile project binary (from project root)
  postman              Generate Postman collection from API endpoints
                          --project <name>
                          --output <path>
                          --base-url <url>
  makemigrations [name] Generate migration files from models
  migrate               Apply pending migrations
  migrate --rollback    Rollback last migration
  doctor                Validate configuration
  version               Show version`)
}
