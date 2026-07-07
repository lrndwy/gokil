package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func (qs *QuerySet[T]) loadSelectRelated(results []*T) error {
	for _, relName := range qs.selectRelated {
		relField, ok := qs.meta.FieldByName[relName]
		if !ok {
			return fmt.Errorf("relation %s not found on %s", relName, qs.meta.Name)
		}
		if relField.Relation.Type != RelationBelongsTo {
			continue
		}

		relatedMeta, ok := GetModel(relField.Relation.RelatedModel)
		if !ok {
			return fmt.Errorf("related model %s not registered", relField.Relation.RelatedModel)
		}

		fkCol := relField.Relation.FKColumn
		if fkCol == "" {
			fkCol = toColumnName(relField.Relation.RelatedModel) + "_id"
		}

		ids := collectFKIDs(results, fkCol, relField.Name)
		if len(ids) == 0 {
			continue
		}

		related, err := loadRelatedByIDs(qs.ctx, relatedMeta, ids)
		if err != nil {
			return err
		}

		assignBelongsTo(results, relField, related)
	}
	return nil
}

func (qs *QuerySet[T]) loadPrefetchRelated(results []*T) error {
	for _, relName := range qs.prefetchRelated {
		relField, ok := qs.meta.FieldByName[relName]
		if !ok {
			return fmt.Errorf("relation %s not found on %s", relName, qs.meta.Name)
		}

		switch relField.Relation.Type {
		case RelationHasMany, RelationReverse:
			if err := qs.prefetchHasMany(results, relField); err != nil {
				return err
			}
		case RelationManyToMany:
			if err := qs.prefetchM2M(results, relField); err != nil {
				return err
			}
		}
	}
	return nil
}

func (qs *QuerySet[T]) prefetchHasMany(results []*T, relField *FieldMeta) error {
	relatedMeta, ok := GetModel(relField.Relation.RelatedModel)
	if !ok {
		return fmt.Errorf("related model %s not registered", relField.Relation.RelatedModel)
	}

	parentIDs := collectParentIDs(results)
	if len(parentIDs) == 0 {
		return nil
	}

	fkCol := relField.Relation.FKColumn
	if fkCol == "" {
		fkCol = resolveHasManyFKColumn(qs.meta, relField, relatedMeta)
	} else if fm, ok := relatedMeta.FieldByName[fkCol]; ok {
		fkCol = fm.Column
	} else {
		fkCol = toColumnName(normalizeFKFieldName(fkCol))
	}
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return fmt.Errorf("no database in context")
	}

	placeholders := make([]string, len(parentIDs))
	args := make([]any, len(parentIDs))
	for i, id := range parentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	cols := []string{}
	for _, f := range relatedMeta.Fields {
		if !f.IsRelation {
			cols = append(cols, f.Column)
		}
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s IN (%s)",
		strings.Join(cols, ", "),
		relatedMeta.TableName,
		fkCol,
		strings.Join(placeholders, ", "),
	)

	rows, err := conn.QueryContext(qs.ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	grouped := map[int64][]any{}
	for rows.Next() {
		item, parentID, err := scanRelatedRow(rows, relatedMeta, fkCol)
		if err != nil {
			return err
		}
		grouped[parentID] = append(grouped[parentID], item)
	}

	assignHasMany(results, relField, grouped)
	return nil
}

func (qs *QuerySet[T]) prefetchM2M(results []*T, relField *FieldMeta) error {
	through := relField.Relation.ThroughTable
	if through == "" {
		through = toTableName(qs.meta.Name) + "_" + toTableName(relField.Relation.RelatedModel)
	}

	relatedMeta, ok := GetModel(relField.Relation.RelatedModel)
	if !ok {
		return fmt.Errorf("related model %s not registered", relField.Relation.RelatedModel)
	}

	parentIDs := collectParentIDs(results)
	if len(parentIDs) == 0 {
		return nil
	}

	conn := connFromContext(qs.ctx)
	if conn == nil {
		return fmt.Errorf("no database in context")
	}

	srcCol := toColumnName(qs.meta.Name) + "_id"
	dstCol := toColumnName(relField.Relation.RelatedModel) + "_id"

	placeholders := make([]string, len(parentIDs))
	args := make([]any, len(parentIDs))
	for i, id := range parentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	relatedCols := []string{}
	for _, f := range relatedMeta.Fields {
		if !f.IsRelation {
			relatedCols = append(relatedCols, "r."+f.Column)
		}
	}

	query := fmt.Sprintf(
		"SELECT t.%s, %s FROM %s t JOIN %s r ON r.id = t.%s WHERE t.%s IN (%s)",
		srcCol,
		strings.Join(relatedCols, ", "),
		through,
		relatedMeta.TableName,
		dstCol,
		srcCol,
		strings.Join(placeholders, ", "),
	)

	rows, err := conn.QueryContext(qs.ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	grouped := map[int64][]any{}
	for rows.Next() {
		var parentID int64
		item, err := scanM2MRow(rows, relatedMeta, &parentID)
		if err != nil {
			return err
		}
		grouped[parentID] = append(grouped[parentID], item)
	}

	assignHasMany(results, relField, grouped)
	return nil
}

func collectFKIDs[T any](results []*T, _ string, relFieldName string) []int64 {
	ids := []int64{}
	seen := map[int64]struct{}{}
	for _, r := range results {
		v := reflect.ValueOf(r).Elem()
		idField := v.FieldByName(relFieldName + "ID")
		if !idField.IsValid() {
			continue
		}
		if idField.Kind() == reflect.Int64 {
			id := idField.Int()
			if _, ok := seen[id]; !ok && id > 0 {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func collectParentIDs[T any](results []*T) []int64 {
	ids := []int64{}
	for _, r := range results {
		v := reflect.ValueOf(r).Elem()
		id := extractID(v)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func extractID(v reflect.Value) int64 {
	if id := v.FieldByName("ID"); id.IsValid() {
		return id.Int()
	}
	if bm := v.FieldByName("BaseModel"); bm.IsValid() {
		if id := bm.FieldByName("ID"); id.IsValid() {
			return id.Int()
		}
	}
	return 0
}

func loadRelatedByIDs(ctx context.Context, meta *ModelMeta, ids []int64) (map[int64]any, error) {
	conn := connFromContext(ctx)
	if conn == nil {
		return nil, fmt.Errorf("no database in context")
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	cols := []string{}
	for _, f := range meta.Fields {
		if !f.IsRelation {
			cols = append(cols, f.Column)
		}
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id IN (%s)",
		strings.Join(cols, ", "),
		meta.TableName,
		strings.Join(placeholders, ", "),
	)

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int64]any{}
	for rows.Next() {
		instance := reflect.New(meta.ModelType).Interface()
		if err := scanInto(rows, meta, instance); err != nil {
			return nil, err
		}
		v := reflect.ValueOf(instance).Elem()
		id := extractID(v)
		result[id] = instance
	}
	return result, nil
}

func assignBelongsTo[T any](results []*T, relField *FieldMeta, related map[int64]any) {
	for _, r := range results {
		v := reflect.ValueOf(r).Elem()
		fkName := relField.Name + "ID"
		fk := v.FieldByName(fkName)
		if !fk.IsValid() {
			continue
		}
		if item, ok := related[fk.Int()]; ok {
			field := v.FieldByName(relField.Name)
			if field.IsValid() && field.CanSet() {
				field.Set(reflect.ValueOf(item))
			}
		}
	}
}

func assignHasMany[T any](results []*T, relField *FieldMeta, grouped map[int64][]any) {
	for _, r := range results {
		v := reflect.ValueOf(r).Elem()
		id := extractID(v)
		items, ok := grouped[id]
		if !ok {
			continue
		}
		field := v.FieldByName(relField.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		slice := reflect.MakeSlice(field.Type(), len(items), len(items))
		for i, item := range items {
			slice.Index(i).Set(reflect.ValueOf(item).Elem())
		}
		field.Set(slice)
	}
}

func scanRelatedRow(rows *sql.Rows, meta *ModelMeta, fkCol string) (any, int64, error) {
	instance := reflect.New(meta.ModelType).Interface()
	if err := scanInto(rows, meta, instance); err != nil {
		return nil, 0, err
	}
	v := reflect.ValueOf(instance).Elem()
	var parentID int64
	if fm, ok := meta.FieldByColumn[fkCol]; ok {
		if f := v.FieldByName(fm.Name); f.IsValid() {
			parentID = f.Int()
		}
	}
	return instance, parentID, nil
}

func scanM2MRow(rows *sql.Rows, meta *ModelMeta, parentID *int64) (any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	instance := reflect.New(meta.ModelType).Interface()
	v := reflect.ValueOf(instance).Elem()

	dest := make([]any, len(cols))
	dest[0] = parentID

	colIdx := 1
	for _, f := range meta.Fields {
		if f.IsRelation {
			continue
		}
		fieldPtr := v.FieldByName(f.Name)
		if !fieldPtr.IsValid() {
			if bm := v.FieldByName("BaseModel"); bm.IsValid() {
				fieldPtr = bm.FieldByName(f.Name)
			}
		}
		if fieldPtr.IsValid() {
			dest[colIdx] = fieldPtr.Addr().Interface()
			colIdx++
		}
	}

	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}
	return instance, nil
}

func scanInto(rows *sql.Rows, meta *ModelMeta, instance any) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	v := reflect.ValueOf(instance).Elem()
	dest := make([]any, len(cols))
	colMap := map[string]int{}
	for i, c := range cols {
		colMap[c] = i
	}
	for _, f := range meta.Fields {
		if f.IsRelation {
			continue
		}
		idx, ok := colMap[f.Column]
		if !ok {
			continue
		}
		fieldPtr := v.FieldByName(f.Name)
		if !fieldPtr.IsValid() {
			if bm := v.FieldByName("BaseModel"); bm.IsValid() {
				fieldPtr = bm.FieldByName(f.Name)
			}
		}
		if fieldPtr.IsValid() {
			dest[idx] = fieldPtr.Addr().Interface()
		}
	}
	return rows.Scan(dest...)
}
