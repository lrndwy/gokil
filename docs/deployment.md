## Deployment & Docker

### Jalankan dengan Docker Compose (dev)

Jika saat `startproject` kamu memilih setup DB/Redis, project akan punya:
- `docker-compose.yml`
- `.env` (generated)

Jalankan:

```bash
docker compose up -d
```

Lalu jalankan server:

```bash
go run ./cmd/<project> serve
```

### Production (ringkas)

Rekomendasi minimal:
- set env `GOKIL_ENV=production`
- set `GOKIL_DEBUG=false`
- set `GOKIL_DB_DSN` sesuai DB production
- jalankan migrasi sebelum start:

```bash
./myapi migrate
./myapi serve
```

