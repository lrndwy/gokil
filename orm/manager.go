package orm

import "context"

func Objects[T any](ctx context.Context) *QuerySet[T] {
	meta, err := ModelMetaFor[T]()
	if err != nil {
		panic(err)
	}
	return &QuerySet[T]{
		ctx:  ctx,
		meta: meta,
	}
}

func Create[T any](ctx context.Context, instance *T) (*T, error) {
	meta, err := ModelMetaFor[T]()
	if err != nil {
		return nil, err
	}
	qs := &QuerySet[T]{ctx: ctx, meta: meta}
	return qs.Create(instance)
}
