package orm

import (
	"reflect"
	"strings"
	"sync"
)

// projections tracks which model fields were loaded via Only(), keyed by instance pointer.
var projections sync.Map // uintptr -> []string (Go field names)

// SetProjection records the fields that were selected for an instance.
func SetProjection(instance any, fields []string) {
	rv := reflect.ValueOf(instance)
	if rv.Kind() != reflect.Ptr || rv.IsNil() || len(fields) == 0 {
		return
	}
	cp := append([]string(nil), fields...)
	projections.Store(rv.Pointer(), cp)
}

// GetProjection returns the Only() fields for an instance, if any.
func GetProjection(instance any) ([]string, bool) {
	rv := reflect.ValueOf(instance)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, false
	}
	v, ok := projections.Load(rv.Pointer())
	if !ok {
		return nil, false
	}
	fields, _ := v.([]string)
	return fields, len(fields) > 0
}

// ClearProjection removes projection metadata for an instance.
func ClearProjection(instance any) {
	rv := reflect.ValueOf(instance)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	projections.Delete(rv.Pointer())
}

// ProjectForJSON converts values so that models loaded with Only() serialize
// only the selected fields (plus always-included PK). Values without a
// projection are returned unchanged for normal encoding/json behavior.
func ProjectForJSON(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out[i] = ProjectForJSON(rv.Index(i).Interface())
		}
		return out
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		if fields, ok := GetProjection(rv.Interface()); ok {
			return structToProjectionMap(rv.Elem(), fields)
		}
		return v
	default:
		return v
	}
}

func structToProjectionMap(v reflect.Value, fields []string) map[string]any {
	wanted := make(map[string]bool, len(fields))
	for _, f := range fields {
		wanted[f] = true
	}
	out := make(map[string]any)
	collectProjectedFields(v, wanted, out)
	return out
}

func collectProjectedFields(v reflect.Value, wanted map[string]bool, out map[string]any) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := v.Field(i)
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			collectProjectedFields(fv, wanted, out)
			continue
		}
		if sf.Anonymous && fv.Kind() == reflect.Ptr && !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
			collectProjectedFields(fv.Elem(), wanted, out)
			continue
		}
		if !wanted[sf.Name] {
			continue
		}
		key := jsonFieldName(sf)
		out[key] = projectFieldValue(fv)
	}
}

func projectFieldValue(fv reflect.Value) any {
	if !fv.IsValid() {
		return nil
	}
	if isBelongsToType(fv.Type()) {
		id := fv.FieldByName("ID")
		ref := fv.FieldByName("Ref")
		m := map[string]any{}
		if id.IsValid() {
			m["ID"] = id.Interface()
		}
		if ref.IsValid() && !ref.IsNil() {
			m["Ref"] = ProjectForJSON(ref.Interface())
		} else {
			m["Ref"] = nil
		}
		return m
	}
	if isHasManyType(fv.Type()) || isManyManyType(fv.Type()) {
		items := fv.FieldByName("Items")
		if !items.IsValid() || items.IsNil() {
			return map[string]any{"Items": []any{}}
		}
		return map[string]any{"Items": ProjectForJSON(items.Interface())}
	}
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return nil
		}
		return ProjectForJSON(fv.Interface())
	}
	return fv.Interface()
}

func jsonFieldName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return sf.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return sf.Name
	}
	return name
}
