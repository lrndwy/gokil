## Error handling & response envelope

### Response sukses

Semua response sukses memakai envelope:

```json
{
  "status": 200,
  "message": "users retrieved",
  "data": [...]
}
```

### Error

Error mengikuti format sederhana:

```json
{ "error": "user not found" }
```

### Helper error

Gunakan helper:
- `views.BadRequest("...")`
- `views.NotFound("...")`
- `views.Unauthorized("...")`
- `views.Forbidden("...")`
- `views.Conflict("...")`
- `views.UnprocessableEntity("...")`
- `views.Validation(map[string]string{...})`

Dan untuk mapping `sql.ErrNoRows`:

```go
if err := views.NotFoundIf(err, "user not found"); err != nil {
    return err
}
```

