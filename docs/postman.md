---
title: Postman Collection
nav_order: 8
---

## Postman Collection Generator

Gokil dapat menghasilkan **Postman Collection v2.1.0** secara otomatis dari source code project. Parser mengekstrak route, request body, path variables, dan query parameters dari file `urls.go` dan `views/*.go`.

## Cara Pakai

Jalankan dari root project (yang punya folder `cmd/<project>`):

```bash
gokil postman
```

### Opsi

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--project` | auto-detect | Nama project (`cmd/<project>`) |
| `--output` | `collection_postman.json` | Path output file |
| `--base-url` | `http://localhost:8080` | Base URL untuk semua request |

### Contoh

```bash
# Basic — auto-detect project name
gokil postman

# Dengan custom output
gokil postman --output ./docs/api-collection.json

# Dengan base URL production
gokil postman --base-url https://api.myapp.com

# Tentukan project name manual
gokil postman --project myapi
```

## Apa yang Di-generate

### Routes

Semua route yang terdaftar di `urls.go` akan di-parse otomatis:

```go
// urls.go
func URLPatterns(app *framework.App, r *router.Router) {
    r.GET("/api/users/", app.Wrap(views.UserList))
    r.POST("/api/users/", app.Wrap(views.UserCreate))
    r.GET("/api/users/:id", app.Wrap(views.UserDetail))
    r.PUT("/api/users/:id", app.Wrap(views.UserUpdate))
    r.DELETE("/api/users/:id", app.Wrap(views.UserDelete))
}
```

Akan menghasilkan endpoint di collection:
- `GET /api/users/`
- `POST /api/users/`
- `GET /api/users/:id`
- `PUT /api/users/:id`
- `DELETE /api/users/:id`

### Request Body (POST/PUT/PATCH)

Parser mengekstrak struct input dengan JSON tags dari view functions:

```go
func UserCreate(ctx *views.Context) error {
    var input struct {
        Email string `json:"email"`
        Name  string `json:"name"`
    }
    // ...
}
```

Akan menghasilkan body JSON di Postman:

```json
{
  "email": "user@example.com",
  "name": "string"
}
```

Type inference:
- `string` → `"string"` (atau `"user@example.com"` untuk field email)
- `int64` / `int` → `1`
- `float64` → `0.0`
- `bool` → `true`

### Path Variables

Parameter URL seperti `:id` otomatis terdeteksi:

```
GET {{base_url}}/api/users/:id
```

Path variables akan ditambahkan ke collection:
```json
{
  "key": "id",
  "value": ":id",
  "description": "Path parameter: id"
}
```

### Query Parameters

Panggilan `ctx.Query()` dan `ctx.QueryInt()` di view functions akan terdeteksi:

```go
func UserList(ctx *views.Context) error {
    page := ctx.QueryInt("page", 1)
    search := ctx.Query("search")
    // ...
}
```

Query parameters ditambahkan ke collection:
```json
{
  "key": "page",
  "value": "0",
  "description": "Query parameter (integer)"
},
{
  "key": "search",
  "value": "",
  "description": "Query parameter"
}
```

### Headers

Setiap request otomatis mendapat headers default:

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `Authorization` | `Bearer {{token}}` |

### Folder Grouping

Routes dikelompokkan otomatis berdasarkan resource:

| Path | Folder |
|------|--------|
| `/api/users/` | Users |
| `/api/posts/` | Posts |
| `/api/tags/` | Tags |
| `/healthz` | Healthz |

### Collection Variables

Collection dilengkapi dengan variabel yang bisa diubah:

```json
{
  "variable": [
    {"key": "base_url", "value": "http://localhost:8080"},
    {"key": "token", "value": "your-jwt-token"}
  ]
}
```

Ubah nilai `token` di Postman setelah import untuk testing dengan auth.

## Import ke Postman

1. Buka Postman
2. Klik **Import** (kiri atas)
3. Pilih **File** → pilih `collection_postman.json`
4. Collection akan muncul di sidebar
5. Ubah variabel `token` di collection settings untuk testing authenticated endpoints

## Struktur Output

```json
{
  "info": {
    "name": "myapi",
    "description": "Auto-generated Postman collection for myapi",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Users",
      "item": [
        {
          "name": "Users",
          "request": {
            "method": "GET",
            "header": [...],
            "url": { "raw": "{{base_url}}/api/users", ... }
          }
        },
        {
          "name": "Create Users",
          "request": {
            "method": "POST",
            "header": [...],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"user@example.com\",\n  \"name\": \"string\"\n}"
            },
            "url": { "raw": "{{base_url}}/api/users", ... }
          }
        }
      ]
    }
  ],
  "variable": [
    {"key": "base_url", "value": "http://localhost:8080"},
    {"key": "token", "value": "your-jwt-token"}
  ]
}
```

## Limitasi

- Parser menggunakan **static analysis** (regex), bukan Go AST parsing
- Hanya mendeteksi pattern `var input struct { ... }` untuk request body
- Tidak mendeteksi body dari `map[string]any` atau variabel non-struct
- Tidak mendeteksi authentication middleware (semua endpoint diasumsikan butuh Bearer token)
- Handler yang menggunakan `app.Wrap()` harus mengikuti pola `views.HandlerName` di `urls.go`
