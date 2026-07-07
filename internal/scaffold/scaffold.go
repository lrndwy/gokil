package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lrndwy/gokil/version"
)

type Options struct {
	Name        string
	Dir         string
	ModPath     string
	ReplacePath string
}

func Create(opts Options) error {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	dir := opts.Dir
	if dir == "" {
		dir = name
	}
	modPath := opts.ModPath
	if modPath == "" {
		modPath = name
	}
	replacePath := opts.ReplacePath
	if replacePath == "" {
		replacePath = ".."
	}
	frameworkVersion := version.RequireVersion()
	modRequire := frameworkVersion
	if modRequire == "latest" {
		modRequire = "v0.0.0"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"go.mod":              goModTemplate,
		"settings.go":         settingsTemplate,
		"models/models.go":    modelsTemplate,
		"urls.go":             urlsTemplate,
		"views/post.go":       viewsPostTemplate,
		"views/user.go":       viewsUserTemplate,
		".env.example":        envExampleTemplate,
		".gitignore":          gitignoreTemplate,
		filepath.Join("cmd", name, "main.go"): mainTemplate,
	}

	for path, tmpl := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := renderTemplate(fullPath, tmpl, map[string]string{
			"Name":             name,
			"ModPath":          modPath,
			"ReplacePath":      replacePath,
			"FrameworkVersion": modRequire,
		}); err != nil {
			return err
		}
	}

	migrationsDir := filepath.Join(dir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return err
	}
	storageDir := filepath.Join(dir, "storage")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return err
	}

	if err := tidyModule(dir, frameworkVersion); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	fmt.Printf("Created project %s\n", dir)
	fmt.Println("Next: cd", dir, "&& cp .env.example .env")
	return nil
}

func tidyModule(dir, frameworkVersion string) error {
	if frameworkVersion == "latest" {
		get := exec.Command("go", "get", version.ModulePath+"@latest")
		get.Dir = dir
		if out, err := get.CombinedOutput(); err != nil {
			return fmt.Errorf("go get latest: %w: %s", err, string(out))
		}
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

func renderTemplate(path, tmpl string, data map[string]string) error {
	t, err := template.New("scaffold").Parse(tmpl)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

const goModTemplate = `module {{.ModPath}}

go 1.22

require github.com/lrndwy/gokil {{.FrameworkVersion}}
`

const settingsTemplate = `package {{.Name}}

import "github.com/lrndwy/gokil/config"

func LoadSettings() (config.Settings, error) {
	return config.Load(config.Options{})
}
`

const modelsTemplate = `package models

import "github.com/lrndwy/gokil/orm"

func init() {
	_ = orm.RegisterModels(
		&User{},
		&Post{},
		&Tag{},
	)
}

type User struct {
	orm.BaseModel
	Email string ` + "`" + `orm:"unique;not null;size:255"` + "`" + `
	Name  string ` + "`" + `orm:"size:100"` + "`" + `
	Posts []Post ` + "`" + `orm:"reverse:author"` + "`" + `
}

type Post struct {
	orm.BaseModel
	Title    string ` + "`" + `orm:"not null;size:200"` + "`" + `
	Content  string ` + "`" + `orm:"type:text"` + "`" + `
	AuthorID int64  ` + "`" + `orm:"not null"` + "`" + `
	Author   *User  ` + "`" + `orm:"fk:AuthorID;rel:belongs_to"` + "`" + `
	Tags     []Tag  ` + "`" + `orm:"m2m:post_tags"` + "`" + `
}

type Tag struct {
	orm.BaseModel
	Name string ` + "`" + `orm:"unique;not null;size:50"` + "`" + `
}
`

const urlsTemplate = `package {{.Name}}

import (
	"{{.ModPath}}/views"
	"github.com/lrndwy/gokil/framework"
	"github.com/lrndwy/gokil/router"
)

func URLPatterns(app *framework.App, r *router.Router) {
	r.GET("/api/health/", app.Wrap(views.HealthCheck))
	r.GET("/api/users/", app.Wrap(views.UserList))
	r.POST("/api/users/", app.Wrap(views.UserCreate))
	r.GET("/api/users/:id", app.Wrap(views.UserDetail))
	r.GET("/api/posts/", app.Wrap(views.PostList))
	r.POST("/api/posts/", app.Wrap(views.PostCreate))
	r.GET("/api/posts/:id", app.Wrap(views.PostDetail))
}
`

const viewsUserTemplate = `package views

import (
	"database/sql"
	"net/http"

	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func HealthCheck(ctx *views.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func UserList(ctx *views.Context) error {
	users, err := orm.Objects[models.User](ctx.DBContext()).All()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, users)
}

func UserCreate(ctx *views.Context) error {
	var input struct {
		Email string ` + "`" + `json:"email"` + "`" + `
		Name  string ` + "`" + `json:"name"` + "`" + `
	}
	if err := ctx.BindJSON(&input); err != nil {
		return views.Error(ctx, http.StatusBadRequest, "invalid json")
	}
	user, err := orm.Create(ctx.DBContext(), &models.User{
		Email: input.Email,
		Name:  input.Name,
	})
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, user)
}

func UserDetail(ctx *views.Context) error {
	user, err := orm.Objects[models.User](ctx.DBContext()).
		Filter("id", ctx.Param("id")).
		Get()
	if err == sql.ErrNoRows {
		return views.Error(ctx, http.StatusNotFound, "user not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, user)
}
`

const viewsPostTemplate = `package views

import (
	"database/sql"
	"net/http"

	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func PostList(ctx *views.Context) error {
	posts, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		All()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, posts)
}

func PostCreate(ctx *views.Context) error {
	var input struct {
		Title    string ` + "`" + `json:"title"` + "`" + `
		Content  string ` + "`" + `json:"content"` + "`" + `
		AuthorID int64  ` + "`" + `json:"author_id"` + "`" + `
	}
	if err := ctx.BindJSON(&input); err != nil {
		return views.Error(ctx, http.StatusBadRequest, "invalid json")
	}
	post, err := orm.Create(ctx.DBContext(), &models.Post{
		Title:    input.Title,
		Content:  input.Content,
		AuthorID: input.AuthorID,
	})
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, post)
}

func PostDetail(ctx *views.Context) error {
	post, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		PrefetchRelated("Tags").
		Filter("id", ctx.Param("id")).
		Get()
	if err == sql.ErrNoRows {
		return views.Error(ctx, http.StatusNotFound, "post not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, post)
}
`

const mainTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"{{.ModPath}}"
	_ "{{.ModPath}}/models"
	"github.com/lrndwy/gokil/framework"
	"github.com/lrndwy/gokil/migration"
	"github.com/lrndwy/gokil/orm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: {{.Name}} <serve|doctor|makemigrations|migrate>")
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
	settings, err := {{.Name}}.LoadSettings()
	if err != nil {
		return err
	}

	app, err := framework.New(settings, nil)
	if err != nil {
		return err
	}

	{{.Name}}.URLPatterns(app, app.Router)
	return app.Run(context.Background())
}

func runDoctor() error {
	settings, err := {{.Name}}.LoadSettings()
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

	settings, err := {{.Name}}.LoadSettings()
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

	settings, err := {{.Name}}.LoadSettings()
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
`

const envExampleTemplate = `# Application
GOKIL_APP_NAME={{.Name}}
GOKIL_ENV=development
GOKIL_DEBUG=true
GOKIL_HOST=127.0.0.1
GOKIL_PORT=8080

# Database (PostgreSQL)
GOKIL_DB_DRIVER=postgres
GOKIL_DB_DSN=postgres://user:password@localhost:5432/{{.Name}}?sslmode=disable
GOKIL_DB_MIGRATIONS_DIR=migrations

# Storage (local or s3)
GOKIL_STORAGE_PROVIDER=local
GOKIL_STORAGE_LOCAL_PATH=storage
# GOKIL_STORAGE_PROVIDER=s3
# GOKIL_STORAGE_BUCKET=my-bucket
# GOKIL_STORAGE_REGION=ap-southeast-1
`

const gitignoreTemplate = `.env
storage/
*.exe
`
