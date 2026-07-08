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

