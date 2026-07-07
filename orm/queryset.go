package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type QuerySet[T any] struct {
	ctx            context.Context
	meta           *ModelMeta
	filters        []filterClause
	orderBy        []string
	limit          int
	offset         int
	selectRelated  []string
	prefetchRelated []string
}

type filterClause struct {
	column   string
	operator string
	value    any
}

func (qs *QuerySet[T]) Filter(lookup string, value any) *QuerySet[T] {
	column, op := parseLookup(lookup)
	qs.filters = append(qs.filters, filterClause{column: column, operator: op, value: value})
	return qs
}

func (qs *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T] {
	qs.orderBy = append(qs.orderBy, fields...)
	return qs
}

func (qs *QuerySet[T]) Limit(n int) *QuerySet[T] {
	qs.limit = n
	return qs
}

func (qs *QuerySet[T]) Offset(n int) *QuerySet[T] {
	qs.offset = n
	return qs
}

func (qs *QuerySet[T]) SelectRelated(fields ...string) *QuerySet[T] {
	qs.selectRelated = append(qs.selectRelated, fields...)
	return qs
}

func (qs *QuerySet[T]) PrefetchRelated(fields ...string) *QuerySet[T] {
	qs.prefetchRelated = append(qs.prefetchRelated, fields...)
	return qs
}

func (qs *QuerySet[T]) All() ([]*T, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return nil, fmt.Errorf("no database in context")
	}

	query, args := qs.buildSelect()
	rows, err := conn.QueryContext(qs.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*T{}
	for rows.Next() {
		instance, err := qs.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, instance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(qs.selectRelated) > 0 {
		if err := qs.loadSelectRelated(results); err != nil {
			return nil, err
		}
	}
	if len(qs.prefetchRelated) > 0 {
		if err := qs.loadPrefetchRelated(results); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (qs *QuerySet[T]) Get() (*T, error) {
	limited := *qs
	limited.limit = 1
	results, err := limited.All()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return results[0], nil
}

func (qs *QuerySet[T]) First() (*T, error) {
	return qs.Limit(1).Get()
}

func (qs *QuerySet[T]) Count() (int64, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return 0, fmt.Errorf("no database in context")
	}

	where, args := qs.buildWhere()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", qs.meta.TableName, where)
	var count int64
	if err := conn.QueryRowContext(qs.ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (qs *QuerySet[T]) Create(instance *T) (*T, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return nil, fmt.Errorf("no database in context")
	}

	now := time.Now().UTC()
	setAutoTimestamps(instance, now)

	columns, placeholders, values := qs.buildInsert(instance)
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		qs.meta.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	var id int64
	if err := conn.QueryRowContext(qs.ctx, query, values...).Scan(&id); err != nil {
		return nil, err
	}
	setFieldValue(reflect.ValueOf(instance).Elem(), "ID", id)
	return instance, nil
}

func (qs *QuerySet[T]) Update(values map[string]any) (int64, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return 0, fmt.Errorf("no database in context")
	}

	sets := []string{}
	args := []any{}
	i := 1
	for key, val := range values {
		col := toColumnName(key)
		if fm, ok := qs.meta.FieldByName[key]; ok {
			col = fm.Column
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	if _, ok := values["UpdatedAt"]; !ok {
		sets = append(sets, fmt.Sprintf("updated_at = $%d", i))
		args = append(args, time.Now().UTC())
		i++
	}

	where, whereArgs := qs.buildWhere()
	for _, a := range whereArgs {
		args = append(args, a)
	}

	query := fmt.Sprintf("UPDATE %s SET %s%s", qs.meta.TableName, strings.Join(sets, ", "), where)
	result, err := conn.ExecContext(qs.ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (qs *QuerySet[T]) Delete() (int64, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return 0, fmt.Errorf("no database in context")
	}

	where, args := qs.buildWhere()
	query := fmt.Sprintf("DELETE FROM %s%s", qs.meta.TableName, where)
	result, err := conn.ExecContext(qs.ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (qs *QuerySet[T]) buildSelect() (string, []any) {
	cols := qs.selectColumns()
	where, args := qs.buildWhere()
	query := fmt.Sprintf("SELECT %s FROM %s%s", strings.Join(cols, ", "), qs.meta.TableName, where)

	if len(qs.orderBy) > 0 {
		parts := make([]string, len(qs.orderBy))
		for i, f := range qs.orderBy {
			if strings.HasPrefix(f, "-") {
				parts[i] = toColumnName(strings.TrimPrefix(f, "-")) + " DESC"
			} else {
				parts[i] = toColumnName(f) + " ASC"
			}
		}
		query += " ORDER BY " + strings.Join(parts, ", ")
	}
	if qs.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qs.limit)
	}
	if qs.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qs.offset)
	}
	return query, args
}

func (qs *QuerySet[T]) selectColumns() []string {
	cols := []string{}
	for _, f := range qs.meta.Fields {
		if f.IsRelation {
			continue
		}
		cols = append(cols, f.Column)
	}
	return cols
}

func (qs *QuerySet[T]) buildWhere() (string, []any) {
	if len(qs.filters) == 0 {
		return "", nil
	}

	clauses := []string{}
	args := []any{}
	for i, f := range qs.filters {
		col := f.column
		if fm, ok := qs.meta.FieldByName[f.column]; ok {
			col = fm.Column
		} else if fm, ok := qs.meta.FieldByColumn[f.column]; ok {
			col = fm.Column
		} else {
			col = toColumnName(f.column)
		}

		clause, val := buildFilterClause(col, f.operator, f.value, i+1)
		clauses = append(clauses, clause)
		if val != nil {
			if slice, ok := val.([]any); ok {
				args = append(args, slice...)
			} else {
				args = append(args, val)
			}
		}
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (qs *QuerySet[T]) buildInsert(instance *T) ([]string, []string, []any) {
	v := reflect.ValueOf(instance).Elem()
	columns := []string{}
	placeholders := []string{}
	values := []any{}
	i := 1

	for _, f := range qs.meta.Fields {
		if f.IsRelation || f.PrimaryKey && f.AutoIncrement {
			continue
		}
		if f.AutoNow || f.AutoNowAdd {
			// handled by setAutoTimestamps
		}
		val := fieldValue(v, f.Name)
		if val == nil && f.Nullable {
			columns = append(columns, f.Column)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, nil)
			i++
			continue
		}
		if isZeroValue(val) && f.Nullable {
			continue
		}
		columns = append(columns, f.Column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}
	return columns, placeholders, values
}

func (qs *QuerySet[T]) scanRow(rows *sql.Rows) (*T, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	dest := make([]any, len(cols))
	instance := new(T)
	v := reflect.ValueOf(instance).Elem()

	colMap := map[string]int{}
	for i, c := range cols {
		colMap[c] = i
	}

	for _, f := range qs.meta.Fields {
		if f.IsRelation {
			continue
		}
		idx, ok := colMap[f.Column]
		if !ok {
			continue
		}
		fieldPtr := v.FieldByName(f.Name)
		if !fieldPtr.IsValid() {
			// BaseModel embedded fields
			if f.Name == "ID" || f.Name == "CreatedAt" || f.Name == "UpdatedAt" {
				if bm := v.FieldByName("BaseModel"); bm.IsValid() {
					fieldPtr = bm.FieldByName(f.Name)
				}
			}
		}
		if !fieldPtr.IsValid() || !fieldPtr.CanSet() {
			continue
		}
		dest[idx] = fieldPtr.Addr().Interface()
	}

	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}
	return instance, nil
}

func parseLookup(lookup string) (string, string) {
	parts := strings.Split(lookup, "__")
	if len(parts) == 1 {
		return parts[0], "exact"
	}
	return strings.Join(parts[:len(parts)-1], "__"), parts[len(parts)-1]
}

func buildFilterClause(col, op string, value any, argIndex int) (string, any) {
	switch op {
	case "exact":
		return fmt.Sprintf("%s = $%d", col, argIndex), value
	case "icontains":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER($%d)", col, argIndex), "%" + fmt.Sprint(value) + "%"
	case "contains":
		return fmt.Sprintf("%s LIKE $%d", col, argIndex), "%" + fmt.Sprint(value) + "%"
	case "gt":
		return fmt.Sprintf("%s > $%d", col, argIndex), value
	case "gte":
		return fmt.Sprintf("%s >= $%d", col, argIndex), value
	case "lt":
		return fmt.Sprintf("%s < $%d", col, argIndex), value
	case "lte":
		return fmt.Sprintf("%s <= $%d", col, argIndex), value
	case "in":
		slice := toAnySlice(value)
		placeholders := make([]string, len(slice))
		for i := range slice {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ", ")), slice
	case "isnull":
		if b, ok := value.(bool); ok && b {
			return fmt.Sprintf("%s IS NULL", col), nil
		}
		return fmt.Sprintf("%s IS NOT NULL", col), nil
	default:
		return fmt.Sprintf("%s = $%d", col, argIndex), value
	}
}

func toAnySlice(value any) []any {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice {
		return []any{value}
	}
	result := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}
	return result
}

func setAutoTimestamps(instance any, now time.Time) {
	v := reflect.ValueOf(instance).Elem()
	if bm := v.FieldByName("BaseModel"); bm.IsValid() {
		if ca := bm.FieldByName("CreatedAt"); ca.IsValid() && ca.CanSet() && ca.IsZero() {
			ca.Set(reflect.ValueOf(now))
		}
		if ua := bm.FieldByName("UpdatedAt"); ua.IsValid() && ua.CanSet() {
			ua.Set(reflect.ValueOf(now))
		}
	}
}

func fieldValue(v reflect.Value, name string) any {
	f := v.FieldByName(name)
	if !f.IsValid() {
		if bm := v.FieldByName("BaseModel"); bm.IsValid() {
			f = bm.FieldByName(name)
		}
	}
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil
		}
		return f.Elem().Interface()
	}
	return f.Interface()
}

func setFieldValue(v reflect.Value, name string, value any) {
	f := v.FieldByName(name)
	if !f.IsValid() {
		if bm := v.FieldByName("BaseModel"); bm.IsValid() {
			f = bm.FieldByName(name)
		}
	}
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(value))
	}
}

func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	return reflect.ValueOf(v).IsZero()
}
