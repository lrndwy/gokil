---
title: Helpers
parent: Views
nav_order: 1
---

## Views helpers (konsep konsisten)

Tujuan helper `views` adalah: handler singkat, parameter jelas, dan tidak perlu closure `func(db context.Context) ...` untuk kasus umum.

### Pola yang disarankan

Ada 2 gaya yang sama-sama valid:

- **QuerySet-first**: langsung kirim `orm.Objects(... )` ke helper `views` (paling ringkas untuk ORM QuerySet).
- **Data-first**: ambil data dulu (pakai orm helper / logika custom), lalu respon pakai helper `views` (sesuai konsep “load data → cek → respond”).

#### List

##### Data-first

```go
users, err := orm.Objects[models.User](ctx.DBContext()).All()
if err != nil {
    return err
}
return views.Listed(ctx, users, "users retrieved")
```

##### QuerySet-first

```go
return views.List(ctx, "tags retrieved", orm.Objects[models.Tag](ctx.DBContext()))
```

#### Detail (custom QuerySet)

##### Data-first

```go
user, err := orm.Objects[models.User](ctx.DBContext()).
    Filter("id", ctx.Param("id")).
    Get()
if err := views.NotFoundIf(err, "user not found"); err != nil {
    return err
}
return views.Detailed(ctx, user, "user retrieved")
```

##### QuerySet-first

```go
return views.Detail(ctx, "post", "post not found",
    orm.Objects[models.Post](ctx.DBContext()).
        SelectRelated("Author").
        PrefetchRelated("Tags").
        Filter("id", ctx.Param("id")),
)
```

#### Create

##### Data-first

```go
created, err := orm.Create(ctx.DBContext(), &models.User{Email: input.Email, Name: input.Name})
if err != nil {
    return err
}
return views.Created(ctx, created, "users created")
```

##### QuerySet-first (helper ORM)

```go
return views.Create(ctx, "user", &models.User{Email: input.Email, Name: input.Name})
```

#### Update

##### Data-first

```go
user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
if err := views.NotFoundIf(err, "user not found"); err != nil {
    return err
}

_, err = orm.UpdateByID[models.User](ctx.DBContext(), user.ID, map[string]any{
    "email": input.Email,
    "name":  input.Name,
})
if err != nil {
    return err
}

userUpdated, err := orm.GetByID[models.User](ctx.DBContext(), user.ID)
if err := views.NotFoundIf(err, "user not found"); err != nil {
    return err
}
return views.Updated(ctx, userUpdated, "users updated")
```

##### QuerySet-first (helper ORM)

```go
return views.Update[models.User](ctx, "id", "user", "user not found", map[string]any{
    "email": input.Email,
    "name":  input.Name,
})
```

#### Delete

##### Data-first

```go
user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
if err := views.NotFoundIf(err, "user not found"); err != nil {
    return err
}

_, err = orm.DeleteByID[models.User](ctx.DBContext(), user.ID)
if err != nil {
    return err
}
return views.Deleted(ctx, user, "users deleted")
```

##### QuerySet-first (helper ORM)

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

