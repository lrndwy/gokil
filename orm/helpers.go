package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

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

// SetM2M replaces all M2M relations for a given field with the provided IDs.
// The field name should match the M2M field on the model (e.g. "Tags").
func SetM2M[T any](ctx context.Context, instance *T, field string, ids ...int64) error {
	meta, err := ModelMetaFor[T]()
	if err != nil {
		return err
	}

	fm, ok := meta.FieldByName[field]
	if !ok || !fm.IsRelation || fm.Relation.Type != RelationManyToMany {
		return fmt.Errorf("field %s is not a ManyToMany relation", field)
	}

	through := fm.Relation.ThroughTable
	if through == "" {
		through = defaultThroughTable(meta.Name, fm.Relation.RelatedModel)
	}

	// Get the parent ID
	v := reflect.ValueOf(instance).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		if bm := v.FieldByName("BaseModel"); bm.IsValid() {
			idField = bm.FieldByName("ID")
		}
	}
	if !idField.IsValid() {
		return fmt.Errorf("cannot find ID field")
	}
	parentID := idField.Int()

	conn := connFromContext(ctx)
	if conn == nil {
		return fmt.Errorf("no database in context")
	}

	// Delete existing M2M entries
	srcCol := toTableName(meta.Name) + "_id"
	dstCol := toTableName(fm.Relation.RelatedModel) + "_id"
	delQuery := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", quoteIdent(through), quoteIdent(srcCol))
	if _, err := conn.ExecContext(ctx, delQuery, parentID); err != nil {
		return fmt.Errorf("delete m2m: %w", err)
	}

	if len(ids) == 0 {
		return nil
	}

	// Insert new M2M entries
	var b strings.Builder
	b.WriteString(fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES ", quoteIdent(through), quoteIdent(srcCol), quoteIdent(dstCol)))
	args := make([]any, 0, len(ids)*2)
	for i, id := range ids {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, parentID, id)
	}

	if _, err := conn.ExecContext(ctx, b.String(), args...); err != nil {
		return fmt.Errorf("insert m2m: %w", err)
	}
	return nil
}
