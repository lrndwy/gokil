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
    _ = orm.RegisterModels(
        &User{},
        &Post{},
        &Tag{},
    )
}

type User struct {
    orm.BaseModel
    Email string `orm:"unique,required,size:255"`
    Name  string `orm:"size:100"`
    Posts orm.HasMany[Post]
}

type Post struct {
    orm.BaseModel
    Title   string `orm:"required,size:200"`
    Content string `orm:"text"`
    Author  orm.BelongsTo[User] `orm:"required"`
    Tags    orm.ManyMany[Tag, TablePostTags]
}

type TablePostTags string

type Tag struct {
    orm.BaseModel
    Name string `orm:"unique,required,size:50"`
}
```

Relasi didefinisikan dengan tipe generic `HasMany`, `BelongsTo`, dan `ManyMany`. FK untuk `BelongsTo` otomatis disimpan di field `.ID` (mis. `Author.ID` → kolom `author_id`).

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

Tag bisa dipisah dengan koma atau titik koma: `unique,required,size:255`

#### Field scalar

| Tag | Deskripsi |
|-----|-----------|
| `pk` | Primary key |
| `unique` | Unique constraint |
| `required` / `not null` | NOT NULL |
| `null` | Nullable |
| `size:N` | VARCHAR(N) |
| `text` / `type:text` | Kolom TEXT |
| `index` | Index |
| `default:value` | Default value |

#### Relasi (cara baru — disarankan)

| Pola field | Deskripsi |
|------------|-----------|
| `Author orm.BelongsTo[User]` | Many-to-one; FK di `Author.ID` → `author_id` |
| `Posts orm.HasMany[Post]` | One-to-many; FK dicari di model child |
| `Tags orm.ManyMany[Tag, TablePostTags]` | Many-to-many; tabel join dari nama tipe (`TablePostTags` → `post_tags`) |

Contoh minimal:

```go
type TablePostTags string

type Post struct {
    orm.BaseModel
    Author orm.BelongsTo[User] `orm:"required"`
    Tags   orm.ManyMany[Tag, TablePostTags]
}

type User struct {
    orm.BaseModel
    Posts orm.HasMany[Post]
}
```

Setelah `SelectRelated("Author")`, akses relasi via `post.Author.Ref`. Setelah `PrefetchRelated("Tags")`, akses via `post.Tags.Items`.

Tag opsional pada relasi: `required`, `fk:CustomID`, `through:custom_table`.

#### Relasi (cara lama — masih didukung)

| Pola field | Tag (opsional) | Hasil |
|------------|----------------|-------|
| `Author *User` + `AuthorID int64` | _(kosong)_ | `belongs_to` otomatis |
| `Posts []Post` | _(kosong)_ | `has_many` otomatis |
| `Tags []Tag` | `many_many:post_tags` | Many-to-many |

```go
type Post struct {
    orm.BaseModel
    AuthorID int64  `orm:"required"`
    Author   *User
    Tags     []Tag `orm:"many_many:post_tags"`
}
```

#### Relasi (legacy — masih didukung)

| Tag | Deskripsi |
|-----|-----------|
| `fk:AuthorID` + `rel:belongs_to` | Belongs-to |
| `reverse:author` | Has-many (legacy) |
| `m2m:post_tags` | Many-to-many (legacy) |

#### Aturan auto-infer

1. `Author *User` → cari field `AuthorID` di struct yang sama
2. `Posts []Post` → cari FK di model `Post` yang menunjuk ke parent (mis. `AuthorID`)
3. Jika FK tidak ditemukan pada slice field → diperlakukan sebagai `many_many`
4. Register semua model terkait dalam satu `RegisterModels(...)` agar inferensi antar-model berjalan

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

**views/user.go (generated by `startproject`):**

```go
func UserList(ctx *views.Context) error {
    return views.ListRespond(ctx, "users retrieved", func(db context.Context) ([]*models.User, error) {
        return orm.Objects[models.User](db).All()
    })
}

func UserCreate(ctx *views.Context) error {
    var input struct {
        Email string `json:"email"`
        Name  string `json:"name"`
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
        return orm.Create(db, &models.User{Email: input.Email, Name: input.Name})
    })
}

func UserDetail(ctx *views.Context) error {
    return views.DetailByID[models.User](ctx, "id", "user", "user not found")
}

func UserUpdate(ctx *views.Context) error {
    // MustBindJSON + RequiredFields + UpdateByParam
    return views.UpdateByParam[models.User](ctx, "id", "user", "user not found", map[string]any{
        "email": input.Email,
        "name":  input.Name,
    })
}

func UserDelete(ctx *views.Context) error {
    return views.DeleteByParam[models.User](ctx, "id", "user", "user not found")
}
```

### View helpers

#### Context — request & response

| Helper | Deskripsi |
|--------|-----------|
| `ctx.Param("id")` | Ambil route param |
| `ctx.MustParam("id")` | Route param wajib ada, else 400 |
| `ctx.Query("search")` | Ambil query string |
| `ctx.QueryInt("page", 1)` | Query int dengan fallback |
| `ctx.QueryInt64("id", 0)` | Query int64 dengan fallback |
| `ctx.QueryBool("active", false)` | Query bool dengan fallback |
| `ctx.Pagination(20, 100)` | Baca `page` & `limit`, return `(page, limit, offset)` |
| `ctx.MustBindJSON(&input)` | Parse JSON, otomatis 400 jika invalid |
| `ctx.JSON(status, payload)` | Response JSON mentah |
| `ctx.OK(message, data)` | Envelope sukses 200 |
| `ctx.Created(message, data)` | Envelope sukses 201 |
| `ctx.Paginated(message, data, meta)` | Envelope list + metadata pagination |
| `ctx.ResourceCreated("user", data)` | Pesan otomatis: `"user created"` |
| `ctx.ResourceOK("retrieved", "user", data)` | Pesan otomatis: `"user retrieved"` |
| `ctx.NoContentResponse()` | Response 204 No Content |

#### Validation

| Helper | Deskripsi |
|--------|-----------|
| `views.Required("email", value)` | 400 jika field kosong |
| `views.RequiredFields(map[string]string{...})` | 400 jika ada field kosong |
| `views.NotFoundIf(err, "not found")` | Map `sql.ErrNoRows` → 404 |

#### HTTP errors (return dari handler)

| Helper | Status |
|--------|--------|
| `views.BadRequest(msg)` | 400 |
| `views.Unauthorized(msg)` | 401 |
| `views.Forbidden(msg)` | 403 |
| `views.NotFound(msg)` | 404 |
| `views.Conflict(msg)` | 409 |
| `views.Validation(msg)` / `UnprocessableEntity(msg)` | 422 |
| `views.Internal(msg)` | 500 |

`App.Wrap` otomatis memetakan error di atas ke JSON `{ "error": "..." }`.

#### ORM shortcuts (single record)

| Helper | Deskripsi |
|--------|-----------|
| `views.FetchByID[T](ctx, id, "not found")` | Get by ID + map 404 |
| `views.FetchByIDParam[T](ctx, "id", "not found")` | Get by route param |
| `views.UpdateByIDParam[T](ctx, "id", values, "not found")` | Update by param |
| `views.DeleteByIDParam[T](ctx, "id", "not found")` | Delete by param |
| `views.FetchQuery[T](ctx, queryFn, "not found")` | Get custom queryset + map 404 |
| `views.ListQuery[T](ctx, queryFn)` | All custom queryset |
| `views.ListRespond[T](ctx, message, queryFn)` | List + envelope (empty slice aman) |
| `views.ListRespondPaginated[T](ctx, message, listFn, countFn)` | List paginated + meta |

#### CRUD one-liners

| Helper | Deskripsi |
|--------|-----------|
| `views.DetailByID[T](ctx, "id", "user", "user not found")` | Detail + `"user retrieved"` |
| `views.DetailByQuery[T](ctx, "post", "not found", queryFn)` | Detail custom queryset |
| `views.CreateAndRespond[T](ctx, "user", createFn)` | Create + `"user created"` |
| `views.UpdateByParam[T](ctx, "id", "user", "not found", values)` | Update + `"user updated"` |
| `views.UpdateAndRefresh[T](ctx, "id", "post", "not found", values, refreshFn)` | Update + reload relasi |
| `views.DeleteByParam[T](ctx, "id", "user", "not found")` | Delete + `"user deleted"` |

#### ORM package shortcuts

| Helper | Deskripsi |
|--------|-----------|
| `orm.GetByID[T](ctx, id)` | Ambil satu record |
| `orm.UpdateByID[T](ctx, id, values)` | Update dan return data terbaru |
| `orm.DeleteByID[T](ctx, id)` | Hapus dan return data yang dihapus |

### Response format

Response sukses:

```json
{
  "status": 200,
  "message": "user retrieved",
  "data": { "id": 1, "email": "a@b.com", "name": "Ali" }
}
```

Response paginated:

```json
{
  "status": 200,
  "message": "users retrieved",
  "data": [{ "id": 1, "email": "a@b.com" }],
  "meta": { "total": 42, "page": 1, "limit": 20, "pages": 3 }
}
```

Response error:

```json
{ "error": "user not found" }
```

### Custom queryset (Post dengan relasi)

```go
func PostDetail(ctx *views.Context) error {
    return views.DetailByQuery(ctx, "post", "post not found", func(db context.Context) (*models.Post, error) {
        return orm.Objects[models.Post](db).
            SelectRelated("Author").
            PrefetchRelated("Tags").
            Filter("id", ctx.Param("id")).
            Get()
    })
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
