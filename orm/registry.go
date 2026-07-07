package orm

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	globalRegistry = &Registry{
		models: make(map[string]*ModelMeta),
	}
	registryMu sync.RWMutex
)

type Registry struct {
	models map[string]*ModelMeta
}

func RegisterModels(models ...any) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	for _, model := range models {
		meta, err := parseModelMeta(model)
		if err != nil {
			return fmt.Errorf("register %T: %w", model, err)
		}
		key := meta.Name
		globalRegistry.models[key] = meta
	}

	for _, meta := range globalRegistry.models {
		if err := finalizeModelRelations(meta); err != nil {
			return err
		}
	}
	return nil
}

func GetModel(name string) (*ModelMeta, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	m, ok := globalRegistry.models[name]
	return m, ok
}

func AllModels() []*ModelMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make([]*ModelMeta, 0, len(globalRegistry.models))
	for _, m := range globalRegistry.models {
		result = append(result, m)
	}
	return result
}

func ModelMetaFor[T any]() (*ModelMeta, error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	meta, ok := GetModel(t.Name())
	if !ok {
		return nil, fmt.Errorf("model %s not registered", t.Name())
	}
	return meta, nil
}

func ResetRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	globalRegistry.models = make(map[string]*ModelMeta)
}
