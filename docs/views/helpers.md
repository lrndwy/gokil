## Views helpers (konsep konsisten)

Tujuan helper `views` adalah: handler singkat, parameter jelas, dan tidak perlu closure `func(db context.Context) ...` untuk kasus umum.

### Pola yang disarankan

#### List

```go
return views.List(ctx, "tags retrieved", orm.Objects[models.Tag](ctx.DBContext()))
```

#### Detail (custom QuerySet)

```go
return views.Detail(ctx, "post", "post not found",
    orm.Objects[models.Post](ctx.DBContext()).
        SelectRelated("Author").
        PrefetchRelated("Tags").
        Filter("id", ctx.Param("id")),
)
```

#### Create

```go
return views.Create(ctx, "user", &models.User{Email: input.Email, Name: input.Name})
```

#### Update

```go
return views.Update[models.User](ctx, "id", "user", "user not found", map[string]any{
    "email": input.Email,
    "name":  input.Name,
})
```

#### Delete

```go
return views.Delete[models.User](ctx, "id", "user", "user not found")
```

#### Paginated

```go
return views.Paginated(ctx, "users retrieved",
    orm.Objects[models.User](ctx.DBContext()),
    20, 100,
)
```

### Helper yang masih ada (legacy)

Helper lama seperti `ListRespond`, `CreateAndRespond`, `DetailByQuery` tetap ada untuk kompatibilitas, tapi disarankan pakai helper baru di atas agar konsisten dan lebih simpel.

