# Gokil

Framework Go backend REST API dengan polan yang mudah dipahami dan mudah digunakan.
Gokil di desain untuk memudahkan pengembangan API RESTful dengan Go. Mengambil referensi dari Django, menggunakan bahasa golang, membuat struktur project yang mudah dipahami dan mudah dikembangkan.



## Fitur

- **Settings terpusat** — konfigurasi via `settings.go`, override dengan environment variable `GOKIL_*`
- **URLs** — definisi route di `urls.go`
- **Views** — logika REST API di package `views/`
- **Models terpusat** — satu file `models.go` per project
- **Migrasi** — `makemigrations` dan `migrate` menghasilkan folder `migrations/` seperti Django
- **Storage built-in** — penyimpanan file lokal atau S3/MinIO
- **ORM penuh** — CRUD, filter, relasi `SelectRelated` / `PrefetchRelated`
- **Tanpa tabel bawaan** — tidak ada auto-migration untuk user/auth; semua tabel didefinisikan manual di `models.go`
- **Cron jobs** — cron jobs di `crons/`

## Install

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
```

Setelah terinstall, CLI `gokil` tersedia di `$GOPATH/bin` (pastikan ada di `PATH`).


## Struktur Project

```
myapi/
├── cmd/myapi/main.go    # Entrypoint CLI
├── settings.go          # Konfigurasi project
├── models/models.go     # Semua model (satu file)
├── urls.go              # Definisi route
├── views/               # Handler REST API
├── crons/               # Cron jobs
├── migrations/          # File SQL migrasi
├── docker-compose.yml   # Opsional (jika setup DB/Redis via CLI)
└── storage/             # Upload lokal
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

## Requirements

- Go 1.22+
- PostgreSQL 14+


By Hafi
