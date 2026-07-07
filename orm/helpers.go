package orm

import "context"

// GetByID fetches a single record by primary key.
func GetByID[T any](ctx context.Context, id any) (*T, error) {
	return Objects[T](ctx).Filter("id", id).Get()
}

// DeleteByID removes a record by primary key and returns the deleted row.
func DeleteByID[T any](ctx context.Context, id any) (*T, error) {
	obj, err := GetByID[T](ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := Objects[T](ctx).Filter("id", id).Delete(); err != nil {
		return nil, err
	}
	return obj, nil
}

// UpdateByID updates a record by primary key and returns the updated row.
func UpdateByID[T any](ctx context.Context, id any, values map[string]any) (*T, error) {
	if _, err := GetByID[T](ctx, id); err != nil {
		return nil, err
	}
	if _, err := Objects[T](ctx).Filter("id", id).Update(values); err != nil {
		return nil, err
	}
	return GetByID[T](ctx, id)
}
