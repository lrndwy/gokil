package models

import "github.com/lrndwy/gokil/orm"

type QuerySet[T any] struct {
	qs *orm.QuerySet[T]
}

func Query[T any]() *QuerySet[T] {
	return &QuerySet[T]{
		qs: orm.Objects[T](GetContext()),
	}
}

func (q *QuerySet[T]) Filter(field string, value any) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.Filter(field, value)}
}

func (q *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.OrderBy(fields...)}
}

func (q *QuerySet[T]) Limit(n int) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.Limit(n)}
}

func (q *QuerySet[T]) Offset(n int) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.Offset(n)}
}

func (q *QuerySet[T]) SelectRelated(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.SelectRelated(fields...)}
}

func (q *QuerySet[T]) PrefetchRelated(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.PrefetchRelated(fields...)}
}

// Only restricts SELECT to the given fields (Go name or column name).
// The primary key is always included.
func (q *QuerySet[T]) Only(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{qs: q.qs.Only(fields...)}
}

func (q *QuerySet[T]) All() ([]*T, error) {
	return q.qs.All()
}

func (q *QuerySet[T]) First() (*T, error) {
	return q.qs.Limit(1).Get()
}

func (q *QuerySet[T]) Count() (int64, error) {
	return q.qs.Count()
}

func (q *QuerySet[T]) Delete() error {
	_, err := q.qs.Delete()
	return err
}
