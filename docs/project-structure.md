---
title: Project Structure
nav_order: 4
---

## Struktur project

Project hasil `gokil startproject` umumnya seperti ini:

```
myapi/
├── cmd/myapi/main.go
├── settings.go
├── models/
│   ├── models.go
│   └── helpers.go
├── app/
│   ├── register.go          # generated
│   ├── users/
│   │   ├── route.go
│   │   └── _id/route.go
│   └── posts/
│       ├── route.go
│       └── _id/route.go
├── jobs/cron.go
├── migrations/
├── storage/
├── docker-compose.yml        # opsional
├── .env.example
└── .env                     # opsional (jika generate infra)
```

### `settings.go`

Semua konfigurasi aplikasi dibaca dari `settings.go`, lalu bisa dioverride dengan environment variables `GOKIL_*`.

### `app/`

File-based routing ala Next.js. Path URL diambil dari folder; fungsi `GET`/`POST`/`PUT`/`PATCH`/`DELETE` di `route.go` menjadi handler.

- `app/users/route.go` → `/users`
- `app/users/_id/route.go` → `/users/:id` (`_param` = dynamic segment)

Jalankan `gokil generateroutes` setelah menambah folder route (otomatis juga saat `startproject` / `build`). File `app/register.go` digenerate; jangan diedit manual.

### `models/`

- `models.go` — semua model aplikasi + `orm.RegisterModels(...)` di `init()`
- `helpers.go` — re-export `Query` / `Create` / `Save` / `Delete` dari framework
