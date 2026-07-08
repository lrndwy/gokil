## Struktur project

Project hasil `gokil startproject` umumnya seperti ini:

```
myapi/
├── cmd/myapi/main.go
├── settings.go
├── urls.go
├── models/
│   └── models.go
├── views/
│   ├── user.go
│   ├── post.go
│   └── tag.go
├── migrations/
├── storage/
├── docker-compose.yml        # opsional
├── .env.example
└── .env                     # opsional (jika generate infra)
```

### `settings.go`

Semua konfigurasi aplikasi dibaca dari `settings.go`, lalu bisa dioverride dengan environment variables `GOKIL_*`.

### `urls.go`

Definisi route (URL patterns). Contoh:

```go
r.GET("/api/health/", app.Wrap(views.HealthCheck))
r.GET("/api/users/", app.Wrap(views.UserList))
```

### `views/`

Handler REST API. Disarankan gunakan helper `views` agar kodenya singkat dan konsisten.

### `models/models.go`

Semua model aplikasi, plus `orm.RegisterModels(...)` di `init()`.

