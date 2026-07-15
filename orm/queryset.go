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
	ctx             context.Context
	meta            *ModelMeta
	filters         []filterClause
	orderBy         []string
	limit           int
	offset          int
	only            []string
	selectRelated   []string
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

// Only restricts SELECT to the given model fields (Go name or column name).
// The primary key is always included. Relation names that are BelongsTo resolve
// to their FK column (e.g. Only("Author") → author_id).
func (qs *QuerySet[T]) Only(fields ...string) *QuerySet[T] {
	qs.only = append([]string(nil), fields...)
	return qs
}

func (qs *QuerySet[T]) All() ([]*T, error) {
	conn := connFromContext(qs.ctx)
	if conn == nil {
		return nil, fmt.Errorf("no database in context")
	}

	query, args, err := qs.buildSelect()
	if err != nil {
		return nil, err
	}
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
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", quoteIdent(qs.meta.TableName), where)
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
		quoteIdent(qs.meta.TableName),
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
		sets = append(sets, fmt.Sprintf("%s = $%d", quoteIdent(col), i))
		args = append(args, val)
		i++
	}
	if _, ok := values["UpdatedAt"]; !ok {
		sets = append(sets, fmt.Sprintf("%s = $%d", quoteIdent("updated_at"), i))
		args = append(args, time.Now().UTC())
		i++
	}

	where, whereArgs := qs.buildWhereWithOffset(i - 1)
	for _, a := range whereArgs {
		args = append(args, a)
	}

	query := fmt.Sprintf("UPDATE %s SET %s%s", quoteIdent(qs.meta.TableName), strings.Join(sets, ", "), where)
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
	query := fmt.Sprintf("DELETE FROM %s%s", quoteIdent(qs.meta.TableName), where)
	result, err := conn.ExecContext(qs.ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (qs *QuerySet[T]) buildSelect() (string, []any, error) {
	cols, err := qs.selectColumns()
	if err != nil {
		return "", nil, err
	}
	where, args := qs.buildWhere()
	query := fmt.Sprintf("SELECT %s FROM %s%s", strings.Join(cols, ", "), quoteIdent(qs.meta.TableName), where)

	if len(qs.orderBy) > 0 {
		parts := make([]string, len(qs.orderBy))
		for i, f := range qs.orderBy {
			if strings.HasPrefix(f, "-") {
				parts[i] = quoteIdent(toColumnName(strings.TrimPrefix(f, "-"))) + " DESC"
			} else {
				parts[i] = quoteIdent(toColumnName(f)) + " ASC"
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
	return query, args, nil
}

func (qs *QuerySet[T]) selectColumns() ([]string, error) {
	if len(qs.only) == 0 {
		cols := make([]string, 0, len(qs.meta.Fields))
		for _, f := range qs.meta.Fields {
			if f.IsRelation {
				continue
			}
			cols = append(cols, quoteIdent(f.Column))
		}
		return cols, nil
	}

	wanted := map[string]bool{}
	for _, name := range qs.only {
		fm, err := qs.resolveOnlyField(name)
		if err != nil {
			return nil, err
		}
		wanted[fm.Column] = true
	}

	// Always include primary key so instances remain identifiable.
	for _, f := range qs.meta.Fields {
		if f.PrimaryKey {
			wanted[f.Column] = true
		}
	}

	// Include BelongsTo FK columns required by SelectRelated.
	for _, relName := range qs.selectRelated {
		fm, ok := qs.meta.FieldByName[relName]
		if !ok {
			continue
		}
		if fm.IsRelation && fm.Relation.Type == RelationBelongsTo {
			if fk := qs.belongsToFKField(fm); fk != nil {
				wanted[fk.Column] = true
			}
		}
	}

	cols := make([]string, 0, len(wanted))
	for _, f := range qs.meta.Fields {
		if f.IsRelation {
			continue
		}
		if wanted[f.Column] {
			cols = append(cols, quoteIdent(f.Column))
		}
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("Only(): no selectable columns")
	}
	return cols, nil
}

func (qs *QuerySet[T]) resolveOnlyField(name string) (*FieldMeta, error) {
	if fm, ok := qs.meta.FieldByName[name]; ok {
		if fm.IsRelation {
			if fm.Relation.Type == RelationBelongsTo {
				if fk := qs.belongsToFKField(fm); fk != nil {
					return fk, nil
				}
			}
			return nil, fmt.Errorf("Only(): cannot select relation field %q", name)
		}
		return fm, nil
	}
	if fm, ok := qs.meta.FieldByColumn[name]; ok {
		if fm.IsRelation {
			return nil, fmt.Errorf("Only(): cannot select relation field %q", name)
		}
		return fm, nil
	}
	// Allow snake_case of a Go field name even if column map key differs.
	col := toColumnName(name)
	if fm, ok := qs.meta.FieldByColumn[col]; ok && !fm.IsRelation {
		return fm, nil
	}
	return nil, fmt.Errorf("Only(): unknown field %q on %s", name, qs.meta.Name)
}

func (qs *QuerySet[T]) belongsToFKField(rel *FieldMeta) *FieldMeta {
	if rel == nil {
		return nil
	}
	for i := range qs.meta.Fields {
		f := &qs.meta.Fields[i]
		if f.VirtualFK && f.RelationOwner == rel.Name {
			return f
		}
	}
	if fkName := rel.Relation.FKColumn; fkName != "" {
		if fm, ok := qs.meta.FieldByName[fkName]; ok {
			return fm
		}
		if fm, ok := qs.meta.FieldByColumn[toColumnName(fkName)]; ok {
			return fm
		}
	}
	return nil
}

func (qs *QuerySet[T]) buildWhere() (string, []any) {
	return qs.buildWhereWithOffset(0)
}

func (qs *QuerySet[T]) buildWhereWithOffset(offset int) (string, []any) {
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

		argIndex := offset + i + 1
		clause, val := buildFilterClause(quoteIdent(col), f.operator, f.value, argIndex)
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
			columns = append(columns, quoteIdent(f.Column))
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, nil)
			i++
			continue
		}
		if isZeroValue(val) && f.Nullable {
			continue
		}
		columns = append(columns, quoteIdent(f.Column))
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
		if ptr, ok := bindFieldDest(v, f); ok {
			dest[idx] = ptr
		}
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
	if id, ok := belongsToFKValue(v, name); ok {
		return id
	}
	f := structFieldValue(v, name)
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

func structFieldValue(v reflect.Value, name string) reflect.Value {
	f := v.FieldByName(name)
	if f.IsValid() {
		return f
	}
	if bm := v.FieldByName("BaseModel"); bm.IsValid() {
		return bm.FieldByName(name)
	}
	return reflect.Value{}
}

func bindFieldDest(v reflect.Value, f FieldMeta) (any, bool) {
	if f.VirtualFK {
		rel := v.FieldByName(f.RelationOwner)
		if !rel.IsValid() || !isBelongsToType(rel.Type()) {
			return nil, false
		}
		id := rel.FieldByName("ID")
		if !id.IsValid() || !id.CanSet() {
			return nil, false
		}
		return id.Addr().Interface(), true
	}

	fieldPtr := structFieldValue(v, f.Name)
	if !fieldPtr.IsValid() || !fieldPtr.CanSet() {
		return nil, false
	}
	return fieldPtr.Addr().Interface(), true
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
