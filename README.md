# Gokil v2

Framework Go backend REST API dengan pola pengembangan mirip Django.

## Fitur

- **Settings terpusat** — konfigurasi via `settings.go`, override dengan environment variable `GOKIL_*`
- **URLs** — definisi route di `urls.go`
- **Views** — logika REST API di package `views/`
- **Models terpusat** — satu file `models.go` per project
- **Migrasi** — `makemigrations` dan `migrate` menghasilkan folder `migrations/` seperti Django
- **Storage built-in** — penyimpanan file lokal atau S3/MinIO
- **ORM penuh** — CRUD, filter, relasi `SelectRelated` / `PrefetchRelated`
- **Tanpa tabel bawaan** — tidak ada auto-migration untuk user/auth; semua tabel didefinisikan manual di `models.go`

## Install

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
```

Setelah terinstall, CLI `gokil` tersedia di `$GOPATH/bin` (pastikan ada di `PATH`).

## Quick Start

### 1. Buat project baru

```bash
gokil startproject myapi
cd myapi
cp .env.example .env
# go mod tidy dijalankan otomatis oleh startproject
```

Saat `startproject`, CLI akan menanyakan:
- Setup database dengan Docker Compose? (PostgreSQL / MySQL)
- Setup Redis untuk caching dengan Docker Compose?

Jika memilih ya, project akan dibuat dengan `docker-compose.yml` dan `.env` yang sudah selaras.

Non-interaktif:

```bash
gokil startproject myapi --db --db-engine postgres --redis
gokil startproject myapi --db --db-engine mysql --no-redis
```

Jalankan infrastruktur:

```bash
docker compose up -d
```

> **Catatan:** `makemigrations` / `migrate` saat ini dioptimalkan untuk **PostgreSQL**. MySQL didukung untuk koneksi ORM, tetapi generator migrasi masih menggunakan dialek PostgreSQL.

### 2. Konfigurasi environment

```bash
cp .env.example .env
# Edit GOKIL_DB_DSN untuk PostgreSQL
export $(grep -v '^#' .env | xargs)
```

### 3. Definisikan model di `models/models.go`

```go
package models

import "github.com/lrndwy/gokil/orm"

func init() {
    _ = orm.RegisterModels(&User{})
}

type User struct {
    orm.BaseModel
    Email string `orm:"unique;not null;size:255"`
    Name  string `orm:"size:100"`
}
```

### 4. Generate dan jalankan migrasi

```bash
go run ./cmd/myapi makemigrations initial
go run ./cmd/myapi migrate
```

### 5. Jalankan server

```bash
go run ./cmd/myapi serve
```

Health check: `GET http://localhost:8080/healthz`

## Struktur Project

```
myapi/
├── cmd/myapi/main.go    # Entrypoint CLI
├── settings.go          # Konfigurasi project
├── models/models.go     # Semua model (satu file)
├── urls.go              # Definisi route
├── views/               # Handler REST API
├── migrations/          # File SQL migrasi
├── docker-compose.yml   # Opsional (jika setup DB/Redis via CLI)
└── storage/             # Upload lokal
```

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `GOKIL_APP_NAME` | `gokil` | Nama aplikasi |
| `GOKIL_ENV` | `development` | Environment |
| `GOKIL_DEBUG` | `true` | Mode debug |
| `GOKIL_HOST` | `127.0.0.1` | Host server |
| `GOKIL_PORT` | `8080` | Port server |
| `GOKIL_DB_DSN` | — | Connection string database |
| `GOKIL_DB_HOST` | `localhost` | Host database |
| `GOKIL_DB_PORT` | `5432` | Port database |
| `GOKIL_DB_USER` | — | Username database |
| `GOKIL_DB_PASSWORD` | — | Password database |
| `GOKIL_DB_NAME` | — | Nama database |
| `GOKIL_REDIS_ENABLED` | `false` | Aktifkan Redis |
| `GOKIL_REDIS_URL` | — | URL Redis (`redis://host:6379/0`) |
| `GOKIL_REDIS_HOST` | `localhost` | Host Redis |
| `GOKIL_REDIS_PORT` | `6379` | Port Redis |
| `GOKIL_DB_MIGRATIONS_DIR` | `migrations` | Folder migrasi |
| `GOKIL_STORAGE_PROVIDER` | `local` | `local` atau `s3` |
| `GOKIL_STORAGE_LOCAL_PATH` | `storage` | Path storage lokal |

## ORM API

```go
// Create
user, err := orm.Create(ctx, &myapi.User{Email: "a@b.com", Name: "Ali"})

// Query
users, err := orm.Objects[myapi.User](ctx).Filter("name__icontains", "ali").All()
user, err := orm.Objects[myapi.User](ctx).Filter("email", "a@b.com").Get()

// Update
_, err := orm.Objects[myapi.User](ctx).Filter("id", 1).Update(map[string]any{"name": "Budi"})

// Delete
_, err := orm.Objects[myapi.User](ctx).Filter("id", 1).Delete()

// Shortcuts
user, err := orm.GetByID[myapi.User](ctx, 1)
user, err := orm.UpdateByID[myapi.User](ctx, 1, map[string]any{"name": "Budi"})
user, err := orm.DeleteByID[myapi.User](ctx, 1)

// Relasi
posts, err := orm.Objects[myapi.Post](ctx).SelectRelated("Author").All()
posts, err := orm.Objects[myapi.Post](ctx).PrefetchRelated("Tags").All()

// Transaction
err := orm.WithTx(ctx, func(ctx context.Context, tx *orm.Tx) error {
    // ...
    return nil
})
```

### Struct Tag `orm:`

| Tag | Deskripsi |
|-----|-----------|
| `pk` | Primary key |
| `unique` | Unique constraint |
| `not null` | NOT NULL |
| `size:N` | VARCHAR(N) |
| `type:text` | TEXT column |
| `fk:Column` | Foreign key column |
| `rel:belongs_to` | Relasi belongs-to |
| `reverse:field` | Reverse relation (has many) |
| `m2m:table_name` | Many-to-many through table |

### Filter Lookups

`exact` (default), `icontains`, `contains`, `gt`, `gte`, `lt`, `lte`, `in`, `isnull`

Contoh: `Filter("name__icontains", "ali")`, `Filter("id__in", []int64{1,2,3})`

## URLs & Views

**urls.go:**

```go
func URLPatterns(app *framework.App, r *router.Router) {
    r.GET("/api/users/", app.Wrap(views.UserList))
    r.POST("/api/users/", app.Wrap(views.UserCreate))
    r.GET("/api/users/:id", app.Wrap(views.UserDetail))
    r.PUT("/api/users/:id", app.Wrap(views.UserUpdate))
    r.DELETE("/api/users/:id", app.Wrap(views.UserDelete))
}
```

**views/user.go:**

```go
func UserDetail(ctx *views.Context) error {
    user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
    if err := views.NotFoundIf(err, "user not found"); err != nil {
        return err
    }
    return ctx.OK("user retrieved", user)
}

func UserDelete(ctx *views.Context) error {
    user, err := orm.DeleteByID[models.User](ctx.DBContext(), ctx.Param("id"))
    if err := views.NotFoundIf(err, "user not found"); err != nil {
        return err
    }
    return ctx.OK("user deleted", user)
}
```

### View helpers

| Helper | Deskripsi |
|--------|-----------|
| `ctx.MustBindJSON(&input)` | Parse JSON, otomatis 400 jika invalid |
| `ctx.OK(message, data)` | Response sukses `{status, message, data}` |
| `ctx.Created(message, data)` | Response 201 dengan envelope standar |
| `views.NotFoundIf(err, "not found")` | Map `sql.ErrNoRows` ke 404 |
| `views.NotFound/BadRequest/Conflict(msg)` | Return error HTTP dengan status tepat |
| `orm.GetByID[T](ctx, id)` | Ambil satu record by ID |
| `orm.UpdateByID[T](ctx, id, values)` | Update dan kembalikan data terbaru |
| `orm.DeleteByID[T](ctx, id)` | Hapus dan kembalikan data yang dihapus |

Response sukses standar:

```json
{
  "status": 200,
  "message": "user retrieved",
  "data": { "id": 1, "email": "a@b.com", "name": "Ali" }
}
```

Response error:

```json
{ "error": "user not found" }
```

## Storage

```go
// Upload file dari view
err := ctx.Storage.Upload(ctx.Request.Context(), "uploads/file.pdf", reader, size, "application/pdf")

// Dapatkan URL publik
url, err := ctx.Storage.URL("uploads/file.pdf")
```

## CLI Commands

### Framework (`gokil`)

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
gokil startproject <name>   # Buat project baru
gokil doctor                # Validasi konfigurasi (dari project)
gokil version
```

### Project (`cmd/<name>`)

```bash
go run ./cmd/myapi serve
go run ./cmd/myapi doctor
go run ./cmd/myapi makemigrations [name]
go run ./cmd/myapi migrate
go run ./cmd/myapi migrate --rollback
```

## Demo Project

Lihat [`examples/demoapi`](examples/demoapi) untuk contoh lengkap dengan model `User`, `Post`, `Tag`, relasi belongs-to, dan many-to-many.

```bash
cd examples/demoapi
export GOKIL_DB_DSN="postgres://user:pass@localhost:5432/demoapi?sslmode=disable"
go run ./cmd/demoapi makemigrations initial
go run ./cmd/demoapi migrate
go run ./cmd/demoapi serve
```

## Arsitektur

```
CLI → config.Load() → framework.New()
                          ├── orm.Connect()      (PostgreSQL)
                          ├── storage.New()      (local/S3)
                          ├── router + urls.go
                          └── views (per-request Context)
```

## Requirements

- Go 1.22+
- PostgreSQL 14+
