package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"demoapi"
	_ "demoapi/models"
	"gokil/framework"
	"gokil/migration"
	"gokil/orm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: demoapi <serve|doctor|makemigrations|migrate>")
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
	settings, err := demoapi.LoadSettings()
	if err != nil {
		return err
	}

	app, err := framework.New(settings, nil)
	if err != nil {
		return err
	}

	demoapi.URLPatterns(app, app.Router)
	return app.Run(context.Background())
}

func runDoctor() error {
	settings, err := demoapi.LoadSettings()
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

	settings, err := demoapi.LoadSettings()
	if err != nil {
		return err
	}
	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required")
	}

	db, err := orm.Connect(settings.Database.Driver, settings.Database.DSN, settings.Database.MaxOpenConns, settings.Database.MaxIdleConns)
	if err != nil {
		return err
	}
	defer db.Close()

	detector := migration.Detector{DB: db.DB}
	diff, err := detector.Detect()
	if err != nil {
		return err
	}
	if !migration.HasChanges(diff) {
		fmt.Println("No changes detected")
		return nil
	}

	path, err := migration.Generator{Dir: settings.Database.MigrationsDir}.GenerateFromDiff(diff, name)
	if err != nil {
		return err
	}
	fmt.Printf("Created migration: %s\n", path)
	return nil
}

func runMigrate(args []string) error {
	rollback := false
	for _, a := range args {
		if a == "--rollback" {
			rollback = true
		}
	}

	settings, err := demoapi.LoadSettings()
	if err != nil {
		return err
	}
	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required")
	}

	db, err := orm.Connect(settings.Database.Driver, settings.Database.DSN, settings.Database.MaxOpenConns, settings.Database.MaxIdleConns)
	if err != nil {
		return err
	}
	defer db.Close()

	runner := migration.Runner{DB: db.DB, Dir: settings.Database.MigrationsDir}
	if rollback {
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
