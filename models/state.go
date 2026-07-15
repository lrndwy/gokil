package models

import (
	"reflect"
	"strings"
	"sync"
)

var originalStates sync.Map

var relationPrefixes = []string{"BelongsTo", "HasMany", "ManyMany"}

func isRelationField(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	pkgPath := t.PkgPath()
	if !strings.Contains(pkgPath, "/orm") {
		return false
	}
	name := t.Name()
	for _, prefix := range relationPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func StoreOriginalState(obj any) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	state := make(map[string]any)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		if sf.Anonymous {
			continue
		}
		if isRelationField(sf.Type) {
			continue
		}
		state[sf.Name] = v.Field(i).Interface()
	}
	originalStates.Store(reflect.ValueOf(obj).Pointer(), state)
}

func getOriginalState(obj any) map[string]any {
	if state, ok := originalStates.Load(reflect.ValueOf(obj).Pointer()); ok {
		return state.(map[string]any)
	}
	return nil
}

func GetChangedFields(obj any) map[string]any {
	original := getOriginalState(obj)
	if original == nil {
		return getAllFields(obj)
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	changes := make(map[string]any)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() || sf.Anonymous {
			continue
		}
		if isRelationField(sf.Type) {
			continue
		}
		current := v.Field(i).Interface()
		orig, exists := original[sf.Name]
		if !exists || !reflect.DeepEqual(current, orig) {
			changes[sf.Name] = current
		}
	}
	return changes
}

func getAllFields(obj any) map[string]any {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fields := make(map[string]any)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() || sf.Anonymous {
			continue
		}
		if isRelationField(sf.Type) {
			continue
		}
		fields[sf.Name] = v.Field(i).Interface()
	}
	return fields
}

func UpdateOriginalState(obj any) {
	StoreOriginalState(obj)
}

func ClearOriginalState(obj any) {
	originalStates.Delete(reflect.ValueOf(obj).Pointer())
}

func getID(obj any) any {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if id := v.FieldByName("ID"); id.IsValid() {
			return id.Interface()
		}
		if bm := v.FieldByName("BaseModel"); bm.IsValid() && bm.Kind() == reflect.Struct {
			if id := bm.FieldByName("ID"); id.IsValid() {
				return id.Interface()
			}
		}
	}
	return nil
}
