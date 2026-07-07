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
| `GOKIL_DB_DSN` | — | PostgreSQL connection string |
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
}
```

**views/user.go:**

```go
func UserList(ctx *views.Context) error {
    users, err := orm.Objects[myapi.User](ctx.DBContext()).All()
    if err != nil {
        return err
    }
    return ctx.JSON(http.StatusOK, users)
}
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
