---
title: Getting Started
nav_order: 2
---

## Instalasi

```bash
go install github.com/lrndwy/gokil/cmd/gokil@latest
```

Pastikan `$GOPATH/bin` ada di `PATH` agar perintah `gokil` bisa dipanggil.

## Quick Start

### Buat project baru

```bash
gokil startproject myapi
cd myapi
cp .env.example .env
```

Jika ingin setup infra otomatis (Docker Compose DB/Redis):

```bash
gokil startproject myapi --db --db-engine postgres --redis
docker compose up -d
```

### Buat model

Edit `models/models.go` lalu register semua model di `init()`:

```go
package models

import "github.com/lrndwy/gokil/orm"

func init() {
	_ = orm.RegisterModels(&User{}, &Post{}, &Tag{})
}

type User struct {
	orm.BaseModel
	Email string `orm:"unique,required,size:255"`
	Posts orm.HasMany[Post]
}

type Post struct {
	orm.BaseModel
	Title  string `orm:"required,size:200"`
	Author orm.BelongsTo[User] `orm:"required"`
}

type Tag struct {
	orm.BaseModel
	Name string `orm:"unique,required,size:50"`
}
```

### Migrasi

```bash
go run ./cmd/myapi makemigrations initial
go run ./cmd/myapi migrate
```

### Jalankan server

```bash
go run ./cmd/myapi serve
```

Health check default: `GET /api/health/`

