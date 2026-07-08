package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gokil-cli-demo"
	_ "gokil-cli-demo/models"
	"github.com/lrndwy/gokil/cliui"
	"github.com/lrndwy/gokil/framework"
	"github.com/lrndwy/gokil/migration"
	"github.com/lrndwy/gokil/orm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: gokil-cli-demo <serve|doctor|makemigrations|migrate>")
	}

	switch os.Args[1] {
	case "serve":
		if err := runServe(); err != nil {
			log.Fatal(err)
		}
	case "doctor":
		if err := runDoctor(); err != nil {
			log.Fatal(err)
		}
	case "makemigrations":
		if err := runMakeMigrations(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "migrate":
		if err := runMigrate(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func runServe() error {
	settings, err := gokil-cli-demo.LoadSettings()
	if err != nil {
		return err
	}

	app, err := framework.New(settings, nil)
	if err != nil {
		return err
	}

	gokil-cli-demo.URLPatterns(app, app.Router)
	return app.Run(context.Background())
}

func runDoctor() error {
	settings, err := gokil-cli-demo.LoadSettings()
	if err != nil {
		return err
	}
	return settings.Validate()
}

func runMakeMigrations(args []string) error {
	name := "auto"
	if len(args) > 0 {
		name = args[0]
	}

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading settings")

	settings, err := gokil-cli-demo.LoadSettings()
	if err != nil {
		sp.Fail("Loading settings")
		return err
	}
	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required")
	}
	sp.Success("Loaded settings")

	sp.Start("Connecting to database")
	db, err := orm.Connect(settings.Database.Driver, settings.Database.DSN, settings.Database.MaxOpenConns, settings.Database.MaxIdleConns)
	if err != nil {
		sp.Fail("Connecting to database")
		return err
	}
	defer db.Close()
	sp.Success("Connected to database")

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
	path, err := migration.Generator{Dir: settings.Database.MigrationsDir}.GenerateFromDiff(diff, name)
	if err != nil {
		sp.Fail("Generating migration files")
		return err
	}
	sp.Success(fmt.Sprintf("Created migration: %s", path))
	return nil
}

func runMigrate(args []string) error {
	rollback := false
	for _, a := range args {
		if a == "--rollback" {
			rollback = true
		}
	}

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading settings")

	settings, err := gokil-cli-demo.LoadSettings()
	if err != nil {
		sp.Fail("Loading settings")
		return err
	}
	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required")
	}
	sp.Success("Loaded settings")

	sp.Start("Connecting to database")
	db, err := orm.Connect(settings.Database.Driver, settings.Database.DSN, settings.Database.MaxOpenConns, settings.Database.MaxIdleConns)
	if err != nil {
		sp.Fail("Connecting to database")
		return err
	}
	defer db.Close()
	sp.Success("Connected to database")

	runner := migration.Runner{DB: db.DB, Dir: settings.Database.MigrationsDir}
	if rollback {
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
