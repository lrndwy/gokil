## Migrasi

Gokil punya workflow mirip Django:

1. `makemigrations` membaca model yang sudah ter-register
2. generator membuat file SQL di folder `migrations/`
3. `migrate` menjalankan SQL dan mencatat status migrasi

### Generate

```bash
go run ./cmd/<project> makemigrations initial
```

Jika tidak ada perubahan:
- output: `No changes detected`

### Apply

```bash
go run ./cmd/<project> migrate
```

Rollback 1 step:

```bash
go run ./cmd/<project> migrate --rollback
```

### Catatan

Saat ini generator migrasi dioptimalkan untuk PostgreSQL. Koneksi MySQL didukung oleh ORM, tapi SQL migrasi masih PostgreSQL dialect.

