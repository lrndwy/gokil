package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lrndwy/gokil/cliui"
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

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Creating project files")

	files := map[string]string{
		"go.mod":              goModTemplate,
		"settings.go":         settingsTemplate,
		"models/models.go":    modelsTemplate,
		"jobs/cron.go":        cronJobsTemplate,
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
	sp.Success("Created project files")

	sp.Start("Running go mod tidy")
	if err := tidyModule(dir, modCfg); err != nil {
		sp.Fail("Running go mod tidy")
		return fmt.Errorf("go mod tidy: %w", err)
	}
	sp.Success("Finished go mod tidy")

	fmt.Println()
	cliui.Successf("Created project %s", dir)
	if templateData.Infra.NeedsDockerCompose() {
		cliui.Infof("Docker Compose: docker-compose.yml")
		cliui.Infof("Environment: .env (generated from .env.example)")
		fmt.Println()
		fmt.Println(cliui.Bold("Next:"))
		fmt.Println("  cd", dir)
		fmt.Println("  docker compose up -d")
		fmt.Println("  go run ./cmd/"+name, "makemigrations initial")
		fmt.Println("  go run ./cmd/"+name, "migrate")
		fmt.Println("  go run ./cmd/"+name, "serve")
		return nil
	}
	fmt.Println()
	fmt.Println(cliui.Bold("Next:"), "cd", dir, "&& cp .env.example .env")
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
	Posts orm.HasMany[Post]
}

type TablePostTags string

type Post struct {
	orm.BaseModel
	Title   string ` + "`" + `orm:"required,size:200"` + "`" + `
	Content string ` + "`" + `orm:"text"` + "`" + `
	Author  orm.BelongsTo[User] ` + "`" + `orm:"required"` + "`" + `
	Tags    orm.ManyMany[Tag, TablePostTags]
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
	return views.List(ctx, "users retrieved",
		orm.Objects[models.User](ctx.DBContext()).PrefetchRelated("Posts"),
	)
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
	return views.Create(ctx, "user", &models.User{Email: input.Email, Name: input.Name})
}

func UserDetail(ctx *views.Context) error {
	return views.DetailByQuery(ctx, "user", "user not found", func(db context.Context) (*models.User, error) {
		return orm.Objects[models.User](db).
			PrefetchRelated("Posts").
			Filter("id", ctx.Param("id")).
			Get()
	})
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
	return views.Update[models.User](ctx, "id", "user", "user not found", map[string]any{
		"email": input.Email,
		"name":  input.Name,
	})
}

func UserDelete(ctx *views.Context) error {
	return views.Delete[models.User](ctx, "id", "user", "user not found")
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
	return views.List(ctx, "posts retrieved",
		orm.Objects[models.Post](ctx.DBContext()).SelectRelated("Author"),
	)
}

func PostCreate(ctx *views.Context) error {
	var input struct {
		Title    string  ` + "`" + `json:"title"` + "`" + `
		Content  string  ` + "`" + `json:"content"` + "`" + `
		AuthorID int64   ` + "`" + `json:"author_id"` + "`" + `
		TagIDs   []int64 ` + "`" + `json:"tag_ids"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	post := &models.Post{
		Title:   input.Title,
		Content: input.Content,
		Author:  orm.BelongsTo[models.User]{ID: input.AuthorID},
	}
	if err := orm.Create(ctx.DBContext(), post); err != nil {
		return err
	}
	if len(input.TagIDs) > 0 {
		if err := orm.SetM2M(ctx.DBContext(), post, "Tags", input.TagIDs...); err != nil {
			return err
		}
	}
	return ctx.ResourceCreated("post", post)
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
		Title    string  ` + "`" + `json:"title"` + "`" + `
		Content  string  ` + "`" + `json:"content"` + "`" + `
		AuthorID int64   ` + "`" + `json:"author_id"` + "`" + `
		TagIDs   []int64 ` + "`" + `json:"tag_ids"` + "`" + `
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	if err := views.UpdateByIDParam[models.Post](ctx, "id", map[string]any{
		"title":     input.Title,
		"content":   input.Content,
		"author_id": input.AuthorID,
	}, "post not found"); err != nil {
		return err
	}
	post, err := getPost(ctx, ctx.Param("id"))
	if err != nil {
		return err
	}
	if input.TagIDs != nil {
		if err := orm.SetM2M(ctx.DBContext(), post, "Tags", input.TagIDs...); err != nil {
			return err
		}
	}
	return ctx.ResourceOK("updated", "post", post)
}

func PostDelete(ctx *views.Context) error {
	return views.Delete[models.Post](ctx, "id", "post", "post not found")
}
`

const viewsTagTemplate = `package views

import (
	"{{.ModPath}}/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	return views.List(ctx, "tags retrieved", orm.Objects[models.Tag](ctx.DBContext()))
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
	return views.Create(ctx, "tag", &models.Tag{Name: input.Name})
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
	return views.Update[models.Tag](ctx, "id", "tag", "tag not found", map[string]any{
		"name": input.Name,
	})
}

func TagDelete(ctx *views.Context) error {
	return views.Delete[models.Tag](ctx, "id", "tag", "tag not found")
}
`

const mainTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"{{.ModPath}}"
	"{{.ModPath}}/jobs"
	_ "{{.ModPath}}/models"
	"github.com/lrndwy/gokil/cron"
	"github.com/lrndwy/gokil/cliui"
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
	case "cron":
		if err := runCron(); err != nil {
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

func runCron() error {
	settings, err := {{.Name}}.LoadSettings()
	if err != nil {
		return err
	}
	if settings.Database.DSN == "" {
		return fmt.Errorf("GOKIL_DB_DSN is required")
	}

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Connecting to database")
	db, err := orm.Connect(settings.Database.Driver, settings.Database.DSN, settings.Database.MaxOpenConns, settings.Database.MaxIdleConns)
	if err != nil {
		sp.Fail("Connecting to database")
		return err
	}
	defer db.Close()
	sp.Success("Connected to database")

	cliui.Infof("Cron started (Ctrl+C to stop)")
	ctx := orm.WithDB(context.Background(), db)
	return cron.Run(ctx, jobs.CronJobs()...)
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

	sp := cliui.NewSpinner(os.Stdout)
	sp.Start("Loading settings")

	settings, err := {{.Name}}.LoadSettings()
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

	settings, err := {{.Name}}.LoadSettings()
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
`

const cronJobsTemplate = `package jobs

import (
	"context"
	"time"

	"github.com/lrndwy/gokil/cron"
)

// CronJobs is the simplest way to define background jobs.
//
// Jobs run forever until the process is stopped.
func CronJobs() []cron.Job {
	return []cron.Job{
		{
			Name:       "hello",
			Every:      1 * time.Minute,
			RunOnStart: true,
			Run: func(ctx context.Context) error {
				// TODO: write your job here
				_ = ctx
				return nil
			},
		},
	}
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
