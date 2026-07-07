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
	Name    string
	Dir     string
	ModPath string
	Infra   *InfraOptions
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
	modCfg := ResolveGoModule(dir)
	frameworkVersion := modCfg.RequireVersion

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	infraOpts := PromptInfraOptions(name, opts.Infra)
	templateData := TemplateData{
		Name:             name,
		ModPath:          modPath,
		ReplacePath:      modCfg.ReplacePath,
		UseLocalReplace:  modCfg.UseReplace,
		FrameworkVersion: frameworkVersion,
		Infra:            BuildInfraConfig(name, infraOpts),
	}
	if err := validateInfra(templateData); err != nil {
		return err
	}

	files := map[string]string{
		"go.mod":              goModTemplate,
		"settings.go":         settingsTemplate,
		"models/models.go":    modelsTemplate,
		"urls.go":             urlsTemplate,
		"views/post.go":       viewsPostTemplate,
		"views/user.go":       viewsUserTemplate,
		"views/tag.go":        viewsTagTemplate,
		".env.example":        envExampleTemplate,
		".gitignore":          gitignoreTemplate,
		filepath.Join("cmd", name, "main.go"): mainTemplate,
	}

	for path, tmpl := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := renderTemplate(fullPath, tmpl, templateData); err != nil {
			return err
		}
	}

	if templateData.Infra.NeedsDockerCompose() {
		compose, err := RenderDockerCompose(templateData)
		if err != nil {
			return err
		}
		composePath := filepath.Join(dir, "docker-compose.yml")
		if err := os.WriteFile(composePath, []byte(compose), 0o644); err != nil {
			return err
		}
		if err := renderTemplate(filepath.Join(dir, ".env"), envExampleTemplate, templateData); err != nil {
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

	if err := tidyModule(dir, modCfg); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	fmt.Printf("Created project %s\n", dir)
	if templateData.Infra.NeedsDockerCompose() {
		fmt.Println("Docker Compose: docker-compose.yml")
		fmt.Println("Environment: .env (generated from .env.example)")
		fmt.Println("Next:")
		fmt.Println("  cd", dir)
		fmt.Println("  docker compose up -d")
		fmt.Println("  go run ./cmd/"+name, "makemigrations initial")
		fmt.Println("  go run ./cmd/"+name, "migrate")
		fmt.Println("  go run ./cmd/"+name, "serve")
		return nil
	}
	fmt.Println("Next: cd", dir, "&& cp .env.example .env")
	return nil
}

func tidyModule(dir string, modCfg GoModuleConfig) error {
	if !modCfg.UseReplace {
		versionArg := modCfg.RequireVersion
		if versionArg == "latest" || versionArg == "" {
			versionArg = "latest"
		}
		get := exec.Command("go", "get", version.ModulePath+"@"+versionArg)
		get.Dir = dir
		if out, err := get.CombinedOutput(); err != nil {
			return fmt.Errorf("go get %s: %w: %s", versionArg, err, string(out))
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

func renderTemplate(path, tmpl string, data TemplateData) error {
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
{{if .UseLocalReplace}}
replace github.com/lrndwy/gokil => {{.ReplacePath}}
{{end}}
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
	Email string ` + "`" + `orm:"unique,required,size:255"` + "`" + `
	Name  string ` + "`" + `orm:"size:100"` + "`" + `
	Posts []Post
}

type Post struct {
	orm.BaseModel
	Title    string ` + "`" + `orm:"required,size:200"` + "`" + `
	Content  string ` + "`" + `orm:"text"` + "`" + `
	AuthorID int64  ` + "`" + `orm:"required"` + "`" + `
	Author   *User
	Tags     []Tag ` + "`" + `orm:"many_many:post_tags"` + "`" + `
}

type Tag struct {
	orm.BaseModel
	Name string ` + "`" + `orm:"unique,required,size:50"` + "`" + `
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
	r.PUT("/api/users/:id", app.Wrap(views.UserUpdate))
	r.DELETE("/api/users/:id", app.Wrap(views.UserDelete))
	r.GET("/api/posts/", app.Wrap(views.PostList))
	r.POST("/api/posts/", app.Wrap(views.PostCreate))
	r.GET("/api/posts/:id", app.Wrap(views.PostDetail))
	r.PUT("/api/posts/:id", app.Wrap(views.PostUpdate))
	r.DELETE("/api/posts/:id", app.Wrap(views.PostDelete))
	r.GET("/api/tags/", app.Wrap(views.TagList))
	r.POST("/api/tags/", app.Wrap(views.TagCreate))
	r.GET("/api/tags/:id", app.Wrap(views.TagDetail))
	r.PUT("/api/tags/:id", app.Wrap(views.TagUpdate))
	r.DELETE("/api/tags/:id", app.Wrap(views.TagDelete))
}
`

const viewsUserTemplate = `package views

import (
	"context"
	"net/http"

	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func HealthCheck(ctx *views.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func UserList(ctx *views.Context) error {
	return views.ListRespond(ctx, "users retrieved", func(db context.Context) ([]*models.User, error) {
		return orm.Objects[models.User](db).All()
	})
}

func UserCreate(ctx *views.Context) error {
	var input struct {
		Email string ` + "`" + `json:"email"` + "`" + `
		Name  string ` + "`" + `json:"name"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"email": input.Email,
		"name":  input.Name,
	}); err != nil {
		return err
	}
	return views.CreateAndRespond(ctx, "user", func(db context.Context) (*models.User, error) {
		return orm.Create(db, &models.User{
			Email: input.Email,
			Name:  input.Name,
		})
	})
}

func UserDetail(ctx *views.Context) error {
	return views.DetailByID[models.User](ctx, "id", "user", "user not found")
}

func UserUpdate(ctx *views.Context) error {
	var input struct {
		Email string ` + "`" + `json:"email"` + "`" + `
		Name  string ` + "`" + `json:"name"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"email": input.Email,
		"name":  input.Name,
	}); err != nil {
		return err
	}
	return views.UpdateByParam[models.User](ctx, "id", "user", "user not found", map[string]any{
		"email": input.Email,
		"name":  input.Name,
	})
}

func UserDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.User](ctx, "id", "user", "user not found")
}
`

const viewsPostTemplate = `package views

import (
	"context"

	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func getPost(ctx *views.Context, id string) (*models.Post, error) {
	return views.FetchQuery(ctx, func(db context.Context) (*models.Post, error) {
		return orm.Objects[models.Post](db).
			SelectRelated("Author").
			PrefetchRelated("Tags").
			Filter("id", id).
			Get()
	}, "post not found")
}

func PostList(ctx *views.Context) error {
	return views.ListRespond(ctx, "posts retrieved", func(db context.Context) ([]*models.Post, error) {
		return orm.Objects[models.Post](db).SelectRelated("Author").All()
	})
}

func PostCreate(ctx *views.Context) error {
	var input struct {
		Title    string ` + "`" + `json:"title"` + "`" + `
		Content  string ` + "`" + `json:"content"` + "`" + `
		AuthorID int64  ` + "`" + `json:"author_id"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	return views.CreateAndRespond(ctx, "post", func(db context.Context) (*models.Post, error) {
		return orm.Create(db, &models.Post{
			Title:    input.Title,
			Content:  input.Content,
			AuthorID: input.AuthorID,
		})
	})
}

func PostDetail(ctx *views.Context) error {
	return views.DetailByQuery(ctx, "post", "post not found", func(db context.Context) (*models.Post, error) {
		return orm.Objects[models.Post](db).
			SelectRelated("Author").
			PrefetchRelated("Tags").
			Filter("id", ctx.Param("id")).
			Get()
	})
}

func PostUpdate(ctx *views.Context) error {
	var input struct {
		Title    string ` + "`" + `json:"title"` + "`" + `
		Content  string ` + "`" + `json:"content"` + "`" + `
		AuthorID int64  ` + "`" + `json:"author_id"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	return views.UpdateAndRefresh(ctx, "id", "post", "post not found", map[string]any{
		"title":     input.Title,
		"content":   input.Content,
		"author_id": input.AuthorID,
	}, func(db context.Context, id string) (*models.Post, error) {
		return getPost(ctx, id)
	})
}

func PostDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.Post](ctx, "id", "post", "post not found")
}
`

const viewsTagTemplate = `package views

import (
	"context"

	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	return views.ListRespond(ctx, "tags retrieved", func(db context.Context) ([]*models.Tag, error) {
		return orm.Objects[models.Tag](db).All()
	})
}

func TagCreate(ctx *views.Context) error {
	var input struct {
		Name string ` + "`" + `json:"name"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.Required("name", input.Name); err != nil {
		return err
	}
	return views.CreateAndRespond(ctx, "tag", func(db context.Context) (*models.Tag, error) {
		return orm.Create(db, &models.Tag{Name: input.Name})
	})
}

func TagDetail(ctx *views.Context) error {
	return views.DetailByID[models.Tag](ctx, "id", "tag", "tag not found")
}

func TagUpdate(ctx *views.Context) error {
	var input struct {
		Name string ` + "`" + `json:"name"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.Required("name", input.Name); err != nil {
		return err
	}
	return views.UpdateByParam[models.Tag](ctx, "id", "tag", "tag not found", map[string]any{
		"name": input.Name,
	})
}

func TagDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.Tag](ctx, "id", "tag", "tag not found")
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

# Database ({{if eq .Infra.DBDriver "mysql"}}MySQL{{else}}PostgreSQL{{end}})
GOKIL_DB_DRIVER={{.Infra.DBDriver}}
GOKIL_DB_HOST={{.Infra.DBHost}}
GOKIL_DB_PORT={{.Infra.DBPort}}
GOKIL_DB_USER={{.Infra.DBUser}}
GOKIL_DB_PASSWORD={{.Infra.DBPassword}}
GOKIL_DB_NAME={{.Infra.DBName}}
GOKIL_DB_DSN={{.Infra.DBDSN}}
GOKIL_DB_MIGRATIONS_DIR=migrations

{{if .Infra.SetupRedis}}# Redis
GOKIL_REDIS_ENABLED=true
GOKIL_REDIS_HOST={{.Infra.RedisHost}}
GOKIL_REDIS_PORT={{.Infra.RedisPort}}
GOKIL_REDIS_URL={{.Infra.RedisURL}}
{{else}}# Redis (optional)
# GOKIL_REDIS_ENABLED=true
# GOKIL_REDIS_HOST=localhost
# GOKIL_REDIS_PORT=6379
# GOKIL_REDIS_URL=redis://localhost:6379/0
{{end}}
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
