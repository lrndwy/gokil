---
title: QuerySet
parent: ORM
nav_order: 2
---

## QuerySet

QuerySet adalah API query utama:

```go
qs := orm.Objects[models.User](ctx)
```

### Filter

```go
users, err := orm.Objects[models.User](ctx).
    Filter("email", "a@b.com").
    All()
```

Lookup yang tersedia (contoh):
- `name__icontains`
- `age__gt`
- `id__in`
- `deleted_at__isnull`

### OrderBy / Limit / Offset

```go
items, err := orm.Objects[models.User](ctx).
    OrderBy("-id").
    Limit(20).
    Offset(0).
    All()
```

### Get / First

```go
u, err := orm.Objects[models.User](ctx).Filter("id", 1).Get()
```

### Only (partial columns)

Ambil hanya kolom tertentu (lebih hemat jaringan/memori). Field boleh nama Go (`Name`) atau nama kolom (`name`). Primary key selalu ikut diselect.

```go
// ORM rendah
users, err := orm.Objects[models.User](ctx).
    Only("ID", "Name", "Email").
    All()

// Via helpers project (models.Query)
users, err := models.Query[models.User]().
    Only("name", "email").
    All()
```

Catatan:
- Field yang tidak disebut tetap ada di struct, tapi nilainya zero-value.
- `Only("Author")` pada `BelongsTo` di-resolve ke FK (`author_id`).
- Field relasi `HasMany` / `ManyMany` tidak bisa dipakai di `Only`.
- Jika dipakai bersama `SelectRelated`, kolom FK relasi itu juga otomatis ikut diselect.

### SelectRelated / PrefetchRelated

- `SelectRelated("Author")` untuk `BelongsTo` (join/load by ids)
- `PrefetchRelated("Tags")` untuk `HasMany` / `ManyMany`

```go
posts, err := orm.Objects[models.Post](ctx).
    SelectRelated("Author").
    PrefetchRelated("Tags").
    All()
```

Jika model pakai `BelongsTo` generic:
- `post.Author.ID` berisi FK
- `post.Author.Ref` berisi object hasil `SelectRelated`

