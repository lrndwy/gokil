---
title: CLI
nav_order: 3
---

## CLI `gokil`

### `startproject`

Membuat project baru.

```bash
gokil startproject myapi
```

Opsi non-interaktif:

```bash
gokil startproject myapi --db --db-engine postgres --redis
gokil startproject myapi --db --db-engine mysql --no-redis
gokil startproject myapi --no-db --no-redis
```

Output project berisi:
- `settings.go`
- `urls.go`
- `views/`
- `models/models.go`
- `migrations/`
- `.env.example` (dan `.env` jika memilih infra)
- `docker-compose.yml` (opsional)
- `Dockerfile` (opsional, untuk menjalankan app via Docker)

### `compose`

Generate atau update `docker-compose.yml` agar ada service aplikasi Gokil (plus auto-generate `Dockerfile` jika belum ada).

Jalankan dari root project (yang punya folder `cmd/<project>`):

```bash
gokil compose
```

Jika `docker-compose.yml` sudah ada, default-nya akan di-*update* (menambahkan service `gokil` tanpa menghapus service lain).

Opsi:

```bash
gokil compose --service api
gokil compose --out docker-compose.yml --update=true
gokil compose --only-app
gokil compose --project myapi
```

### `build`

Compile project jadi binary (default output `./bin/<project>`).

```bash
gokil build
gokil build -o ./bin/myapi
gokil build --os linux --arch amd64
gokil build --project myapi
```

### `postman`

Generate Postman Collection v2.1.0 dari source code. Parser otomatis mengekstrak route, request body, path variables, dan query parameters dari `urls.go` dan `views/*.go`.

```bash
gokil postman
gokil postman --project myapi
gokil postman --output collection.json
gokil postman --base-url http://localhost:8080
```

Output default: `collection_postman.json` di direktori project.

Opsi:

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--project` | auto-detect | Nama project (`cmd/<project>`) |
| `--output` | `collection_postman.json` | Path output file |
| `--base-url` | `http://localhost:8080` | Base URL untuk semua request |

Lihat [Postman Collection](./postman.md) untuk detail lebih lanjut.

### `makemigrations`

Generate file migrasi dari model.

```bash
go run ./cmd/<project> makemigrations initial
```

### `migrate`

Apply migrasi.

```bash
go run ./cmd/<project> migrate
go run ./cmd/<project> migrate --rollback
```

### `doctor`

Validasi konfigurasi (settings, DB, storage).

```bash
gokil doctor
```

### `version`

```bash
gokil version
```

