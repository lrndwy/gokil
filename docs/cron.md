---
title: Cron Jobs
nav_order: 8
---

## Cron Jobs

Gokil menyediakan runner sederhana untuk background jobs: `github.com/lrndwy/gokil/cron`.

### Cara pakai (project scaffold)

Project baru otomatis punya:
- `jobs/cron.go` → daftar job
- command: `cron`

Jalankan:

```bash
go run ./cmd/<project> cron
```

### Menulis job

Edit `jobs/cron.go`:

```go
return []cron.Job{
  {
    Name: "cleanup",
    Every: 10 * time.Minute,
    RunOnStart: true,
    Run: func(ctx context.Context) error {
      // ctx sudah berisi DB via orm.WithDB(...)
      // contoh: users, err := orm.Objects[models.User](ctx).All()
      return nil
    },
  },
}
```

### Behavior

- Job berjalan terus sampai proses berhenti
- Jika job error: defaultnya hanya di-log (bisa override via `cron.Runner{OnError: ...}`)

