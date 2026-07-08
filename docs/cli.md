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

