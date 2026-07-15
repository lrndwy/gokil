# Gokil

Framework Go backend REST API dengan pola yang mudah dipahami dan mudah digunakan.
Gokil di desain untuk memudahkan pengembangan API RESTful dengan Go. Mengambil referensi dari Next.js (file-based routing), menggunakan bahasa golang, membuat struktur project yang mudah dipahami dan mudah dikembangkan.



## Fitur

- **Settings terpusat** — konfigurasi via `settings.go`, override dengan environment variable `GOKIL_*`
- **File-based routing** — route dari folder `app/**/route.go` (auto-generated `app/register.go`)
- **Views** — logika REST API dengan envelope JSON responses (`ctx.Success`, `ctx.Error`)
- **Models terpusat** — satu file `models.go` per project
- **Migrasi** — `makemigrations` dan `migrate` menghasilkan folder `migrations/` seperti Django
- **Storage built-in** — penyimpanan file lokal atau S3/MinIO
- **ORM penuh** — CRUD, filter, relasi `SelectRelated` / `PrefetchRelated`, diff-based `Save()`
- **Tanpa tabel bawaan** — tidak ada auto-migration untuk user/auth; semua tabel didefinisikan manual di `models.go`
- **Cron jobs** — cron jobs di `jobs/`
- **Postman collection** — generate collection Postman v2.1.0 dari source code (`gokil postman`)

## Install

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
```

Setelah terinstall, CLI `gokil` tersedia di `$GOPATH/bin` (pastikan ada di `PATH`).


## Struktur Project

```
myapi/
├── cmd/myapi/main.go        # Entrypoint CLI
├── settings.go              # Konfigurasi project
├── models/
│   ├── models.go            # Semua model
│   └── helpers.go           # Re-export Query/Create/Save/Delete
├── app/                     # File-based routing
│   ├── register.go          # Generated — jangan diedit
│   ├── users/
│   │   ├── route.go         # GET, POST → /users
│   │   └── _id/route.go     # GET, PUT, DELETE → /users/:id
│   └── posts/
│       ├── route.go
│       └── _id/route.go
├── jobs/cron.go             # Cron jobs
├── migrations/              # File SQL migrasi
├── docker-compose.yml       # Opsional (jika setup DB/Redis via CLI)
└── storage/                 # Upload lokal
```

## File-based Routing

Route ditentukan oleh folder di bawah `app/`. Export fungsi HTTP method (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) di `route.go` — tanpa `RegisterRoute`.

Folder `_id` menjadi dynamic segment `:id` (pengganti `[id]` Next.js, karena Go tidak mengizinkan `[id]` di import path).

```go
// app/users/route.go
package users

import (
    "myapi/models"
    "github.com/lrndwy/gokil/views"
)

func GET(ctx *views.Context) error {
    users, err := models.Query[models.User]().All()
    if err != nil {
        return ctx.Error(500, err.Error())
    }
    return ctx.Success(200, "users retrieved", users)
}

func POST(ctx *views.Context) error {
    var body struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := ctx.Bind(&body); err != nil {
        return ctx.Error(400, err.Error())
    }
    user := &models.User{Name: body.Name, Email: body.Email}
    if err := models.Create(user); err != nil {
        return ctx.Error(500, err.Error())
    }
    return ctx.Success(201, "user created", user)
}
```

Setelah menambah/mengubah folder route, jalankan:

```bash
gokil generateroutes
```

(`gokil build` dan `gokil startproject` juga menjalankan ini otomatis.)

## ORM with Generics

```go
// Query
users, err := models.Query[models.User]().Filter("name", "John").All()
user, err := models.Query[models.User]().Filter("id", id).First()

// Create
err := models.Create(&models.User{Name: "John"})

// Save (diff-based - only updates changed fields)
user.Name = "Jane"
err := models.Save(user)

// Delete
err := models.Delete[models.User](id)
```

## CLI Commands

### Framework (`gokil`)

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
gokil startproject <name>   # Buat project baru
gokil compose               # Generate/update docker-compose.yml + Dockerfile (di project)
gokil build                 # Compile project jadi ./bin/<project>
gokil generateroutes        # Generate app/register.go dari app/**/route.go
gokil postman               # Generate Postman collection dari API endpoints
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

## Requirements

- Go 1.22+
- PostgreSQL 14+


Developed by Hafiz.
