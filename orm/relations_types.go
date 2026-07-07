package orm

// BelongsTo represents a many-to-one relation (e.g. Post belongs to User).
// The FK column defaults to {FieldName}ID (e.g. AuthorID -> author_id).
type BelongsTo[T any] struct {
	ID  int64
	Ref *T
}

// HasMany represents a one-to-many relation (e.g. User has many Posts).
type HasMany[T any] struct {
	Items []*T
}

// ManyMany represents a many-to-many relation through a join table.
// Example: type TablePostTags string; Tags ManyMany[Tag, TablePostTags]
// The join table name is inferred from the table type (TablePostTags -> post_tags).
// Override with tag: `orm:"through:custom_table"`.
type ManyMany[T any, Table ~string] struct {
	Items []*T
}
