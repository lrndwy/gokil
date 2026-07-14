package orm

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type FieldType string

const (
	FieldTypeString    FieldType = "string"
	FieldTypeText      FieldType = "text"
	FieldTypeInteger   FieldType = "integer"
	FieldTypeBigInt    FieldType = "bigint"
	FieldTypeBoolean   FieldType = "boolean"
	FieldTypeFloat     FieldType = "float"
	FieldTypeJSON      FieldType = "json"
	FieldTypeUUID      FieldType = "uuid"
	FieldTypeDateTime  FieldType = "datetime"
	FieldTypeTimestamp FieldType = "timestamp"
)

type RelationType string

const (
	RelationBelongsTo RelationType = "belongs_to"
	RelationHasMany   RelationType = "has_many"
	RelationManyToMany RelationType = "m2m"
	RelationReverse   RelationType = "reverse"
)

type FieldMeta struct {
	Name          string
	Column        string
	GoType        reflect.Type
	FieldType     FieldType
	Size          int
	PrimaryKey    bool
	Unique        bool
	Nullable      bool
	AutoIncrement bool
	AutoNow       bool
	AutoNowAdd    bool
	Default       string
	Index         bool
	IsRelation    bool
	VirtualFK     bool
	RelationOwner string
	Relation      RelationMeta
	StructField   reflect.StructField
}

type RelationMeta struct {
	Type         RelationType
	FKColumn     string
	RelatedModel string
	ThroughTable string
	ReverseField string
}

type ModelMeta struct {
	Name       string
	TableName  string
	ModelType  reflect.Type
	Fields     []FieldMeta
	FieldByName map[string]*FieldMeta
	FieldByColumn map[string]*FieldMeta
	Relations  []FieldMeta
}

func parseModelMeta(model any) (*ModelMeta, error) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct, got %s", t.Kind())
	}

	meta := &ModelMeta{
		Name:          t.Name(),
		TableName:     toTableName(t.Name()),
		ModelType:     t,
		FieldByName:   map[string]*FieldMeta{},
		FieldByColumn: map[string]*FieldMeta{},
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		// Handle embedded BaseModel fields
		if sf.Anonymous && sf.Type == reflect.TypeOf(BaseModel{}) {
			for _, f := range baseModelFields(sf) {
				meta.Fields = append(meta.Fields, f)
				idx := len(meta.Fields) - 1
				meta.FieldByName[f.Name] = &meta.Fields[idx]
				meta.FieldByColumn[f.Column] = &meta.Fields[idx]
			}
			continue
		}

		fm, err := parseFieldMeta(sf)
		if err != nil {
			return nil, err
		}
		if fm == nil {
			continue
		}

		meta.Fields = append(meta.Fields, *fm)
		meta.FieldByName[fm.Name] = &meta.Fields[len(meta.Fields)-1]
		meta.FieldByColumn[fm.Column] = &meta.Fields[len(meta.Fields)-1]
		if fm.IsRelation {
			meta.Relations = append(meta.Relations, *fm)
		}
	}

	return meta, nil
}

func baseModelFields(sf reflect.StructField) []FieldMeta {
	return []FieldMeta{
		{
			Name: "ID", Column: "id", GoType: reflect.TypeOf(int64(0)),
			FieldType: FieldTypeBigInt, PrimaryKey: true, AutoIncrement: true,
			StructField: sf,
		},
		{
			Name: "CreatedAt", Column: "created_at", GoType: reflect.TypeOf(time.Time{}),
			FieldType: FieldTypeTimestamp, AutoNowAdd: true, StructField: sf,
		},
		{
			Name: "UpdatedAt", Column: "updated_at", GoType: reflect.TypeOf(time.Time{}),
			FieldType: FieldTypeTimestamp, AutoNow: true, StructField: sf,
		},
	}
}

func parseFieldMeta(sf reflect.StructField) (*FieldMeta, error) {
	tag := sf.Tag.Get("orm")
	if tag == "-" {
		return nil, nil
	}

	if fm, ok := parseTypedRelation(sf); ok {
		applyFieldTagOptions(tag, fm)
		return fm, nil
	}

	opts := parseTag(tag)
	fm := &FieldMeta{
		Name:        sf.Name,
		Column:      toColumnName(sf.Name),
		GoType:      sf.Type,
		FieldType:   inferFieldType(sf.Type),
		Nullable:    isNullable(sf.Type),
		StructField: sf,
	}

	for _, opt := range opts {
		switch {
		case opt == "pk":
			fm.PrimaryKey = true
		case opt == "unique":
			fm.Unique = true
		case opt == "not null", opt == "required":
			fm.Nullable = false
		case opt == "null":
			fm.Nullable = true
		case opt == "auto":
			fm.AutoIncrement = true
		case opt == "auto_now":
			fm.AutoNow = true
		case opt == "auto_now_add":
			fm.AutoNowAdd = true
		case opt == "index":
			fm.Index = true
		case opt == "text":
			fm.FieldType = FieldTypeText
		case opt == "belongs_to":
			fm.IsRelation = true
			fm.Relation.Type = RelationBelongsTo
			fm.Relation.FKColumn = sf.Name + "ID"
		case strings.HasPrefix(opt, "belongs_to:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationBelongsTo
			fm.Relation.FKColumn = strings.TrimPrefix(opt, "belongs_to:")
		case opt == "has_many":
			fm.IsRelation = true
			fm.Relation.Type = RelationHasMany
		case strings.HasPrefix(opt, "has_many:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationHasMany
			fm.Relation.FKColumn = strings.TrimPrefix(opt, "has_many:")
		case opt == "many_many":
			fm.IsRelation = true
			fm.Relation.Type = RelationManyToMany
		case strings.HasPrefix(opt, "many_many:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationManyToMany
			fm.Relation.ThroughTable = strings.TrimPrefix(opt, "many_many:")
		case strings.HasPrefix(opt, "column:"):
			fm.Column = strings.TrimPrefix(opt, "column:")
		case strings.HasPrefix(opt, "size:"):
			fmt.Sscanf(strings.TrimPrefix(opt, "size:"), "%d", &fm.Size)
		case strings.HasPrefix(opt, "type:"):
			fm.FieldType = FieldType(strings.TrimPrefix(opt, "type:"))
		case strings.HasPrefix(opt, "default:"):
			fm.Default = strings.TrimPrefix(opt, "default:")
		case strings.HasPrefix(opt, "fk:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationBelongsTo
			fm.Relation.FKColumn = strings.TrimPrefix(opt, "fk:")
		case strings.HasPrefix(opt, "rel:"):
			rel := strings.TrimPrefix(opt, "rel:")
			fm.IsRelation = true
			switch {
			case rel == "belongs_to":
				fm.Relation.Type = RelationBelongsTo
			case rel == "has_many":
				fm.Relation.Type = RelationHasMany
			case strings.HasPrefix(rel, "m2m:"):
				fm.Relation.Type = RelationManyToMany
				fm.Relation.ThroughTable = strings.TrimPrefix(rel, "m2m:")
			}
		case strings.HasPrefix(opt, "reverse:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationReverse
			fm.Relation.ReverseField = strings.TrimPrefix(opt, "reverse:")
		case strings.HasPrefix(opt, "m2m:"):
			fm.IsRelation = true
			fm.Relation.Type = RelationManyToMany
			fm.Relation.ThroughTable = strings.TrimPrefix(opt, "m2m:")
		}
	}

	if fm.IsRelation {
		fm.Relation.RelatedModel = relatedModelName(sf.Type)
		return fm, nil
	}

	if sf.Type.Kind() == reflect.Ptr && sf.Type.Elem().Kind() == reflect.Struct && sf.Type != reflect.TypeOf(BaseModel{}) {
		fm.IsRelation = true
		fm.Relation.Type = RelationBelongsTo
		fm.Relation.FKColumn = sf.Name + "ID"
		fm.Relation.RelatedModel = sf.Type.Elem().Name()
		return fm, nil
	}

	if sf.Type.Kind() == reflect.Slice {
		elem := sf.Type.Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct && elem.Name() != "" {
			fm.IsRelation = true
			fm.Relation.Type = RelationHasMany
			fm.Relation.RelatedModel = elem.Name()
			return fm, nil
		}
	}

	return fm, nil
}

func applyFieldTagOptions(tag string, fm *FieldMeta) {
	for _, opt := range parseTag(tag) {
		switch {
		case opt == "not null", opt == "required":
			fm.Nullable = false
		case opt == "null":
			fm.Nullable = true
		case strings.HasPrefix(opt, "fk:"):
			fm.Relation.FKColumn = strings.TrimPrefix(opt, "fk:")
		case strings.HasPrefix(opt, "through:"):
			fm.Relation.ThroughTable = strings.TrimPrefix(opt, "through:")
		case strings.HasPrefix(opt, "column:"):
			fm.Column = strings.TrimPrefix(opt, "column:")
		}
	}
}

func parseTag(tag string) []string {
	if strings.TrimSpace(tag) == "" {
		return nil
	}
	tag = strings.ReplaceAll(tag, ",", ";")
	parts := strings.Split(tag, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func inferFieldType(t reflect.Type) FieldType {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return FieldTypeString
	case reflect.Bool:
		return FieldTypeBoolean
	case reflect.Int, reflect.Int32:
		return FieldTypeInteger
	case reflect.Int64:
		return FieldTypeBigInt
	case reflect.Float32, reflect.Float64:
		return FieldTypeFloat
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return FieldTypeTimestamp
		}
	}
	return FieldTypeString
}

func isNullable(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

func relatedModelName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t.Name()
}

func toTableName(name string) string {
	return ToTableName(name)
}

func ToTableName(name string) string {
	return toSnakeCase(name)
}

func toColumnName(name string) string {
	return toSnakeCase(name)
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			nextLower := i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z'
			if (prev >= 'a' && prev <= 'z') || nextLower {
				b.WriteByte('_')
			}
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func ToColumnName(name string) string {
	return toColumnName(name)
}

func SQLType(f FieldMeta) string {
	switch f.FieldType {
	case FieldTypeString:
		if f.Size > 0 {
			return fmt.Sprintf("VARCHAR(%d)", f.Size)
		}
		return "VARCHAR(255)"
	case FieldTypeText:
		return "TEXT"
	case FieldTypeInteger:
		return "INTEGER"
	case FieldTypeBigInt:
		return "BIGINT"
	case FieldTypeBoolean:
		return "BOOLEAN"
	case FieldTypeFloat:
		return "DOUBLE PRECISION"
	case FieldTypeJSON:
		return "JSONB"
	case FieldTypeUUID:
		return "UUID"
	case FieldTypeDateTime, FieldTypeTimestamp:
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
