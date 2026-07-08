## Storage

Gokil punya abstraction storage:
- local filesystem
- S3 / MinIO (AWS SDK v2)

### Local storage

Default scaffold memakai local storage folder `storage/`.

Env penting:
- `GOKIL_STORAGE_PROVIDER=local`
- `GOKIL_STORAGE_LOCAL_PATH=storage`

### S3 / MinIO

Set:
- `GOKIL_STORAGE_PROVIDER=s3`
- endpoint/credentials sesuai kebutuhan

Catatan: detail field env mengikuti `settings.go` project (lihat file tersebut untuk opsi lengkap).

