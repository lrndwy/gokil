## Model & Relasi

### BaseModel

Tambahkan `orm.BaseModel` agar model punya kolom default:
- `id`
- `created_at`
- `updated_at`

```go
type User struct {
    orm.BaseModel
    Email string `orm:"unique,required,size:255"`
}
```

### Relasi (cara baru — disarankan)

Gokil v2 memakai tipe generic agar relasi mudah dibaca:

- `orm.BelongsTo[T]` (many-to-one)
- `orm.HasMany[T]` (one-to-many)
- `orm.ManyMany[T, TableType]` (many-to-many)

Contoh:

```go
type TablePostTags string

type User struct {
    orm.BaseModel
    Posts orm.HasMany[Post]
}

type Post struct {
    orm.BaseModel
    Author orm.BelongsTo[User] `orm:"required"`
    Tags   orm.ManyMany[Tag, TablePostTags]
}
```

Catatan:
- FK `BelongsTo` disimpan di `Author.ID` (kolom `author_id`).
- Setelah `SelectRelated("Author")` akses objeknya via `post.Author.Ref`.
- Untuk `ManyMany`, nama tabel join diambil dari nama tipe table: `TablePostTags` → `post_tags`.

### Relasi (cara lama — masih didukung)

Masih boleh:

```go
type Post struct {
    orm.BaseModel
    AuthorID int64 `orm:"required"`
    Author   *User
    Tags     []Tag `orm:"many_many:post_tags"`
}
```

