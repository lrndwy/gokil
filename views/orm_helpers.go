package views

import (
	"context"

	"github.com/lrndwy/gokil/orm"
)

// List runs qs.All() and writes the standard success envelope.
//
// Example:
//   return views.List(ctx, "tags retrieved", orm.Objects[models.Tag](ctx.DBContext()))
func List[T any](c *Context, message string, qs *orm.QuerySet[T]) error {
	items, err := qs.All()
	if err != nil {
		return err
	}
	if items == nil {
		items = make([]*T, 0)
	}
	return c.OK(message, items)
}

// Detail runs qs.Get(), maps sql.ErrNoRows to notFound, and writes "<resource> retrieved".
//
// Example:
//   return views.Detail(ctx, "post", "post not found",
//       orm.Objects[models.Post](ctx.DBContext()).SelectRelated("Author").Filter("id", ctx.Param("id")),
//   )
func Detail[T any](c *Context, resource, notFound string, qs *orm.QuerySet[T]) error {
	obj, err := qs.Get()
	if err := NotFoundIf(err, notFound); err != nil {
		return err
	}
	return c.ResourceOK("retrieved", resource, obj)
}

// FetchByID loads one record by id and maps sql.ErrNoRows to notFound.
func FetchByID[T any](c *Context, id any, notFound string) (*T, error) {
	obj, err := orm.GetByID[T](c.DBContext(), id)
	if err := NotFoundIf(err, notFound); err != nil {
		return nil, err
	}
	return obj, nil
}

// FetchByIDParam loads one record using a route param as id.
func FetchByIDParam[T any](c *Context, param, notFound string) (*T, error) {
	return FetchByID[T](c, c.Param(param), notFound)
}

// UpdateByIDParam updates one record using a route param as id.
func UpdateByIDParam[T any](c *Context, param string, values map[string]any, notFound string) (*T, error) {
	obj, err := orm.UpdateByID[T](c.DBContext(), c.Param(param), values)
	if err := NotFoundIf(err, notFound); err != nil {
		return nil, err
	}
	return obj, nil
}

// DeleteByIDParam deletes one record using a route param as id.
func DeleteByIDParam[T any](c *Context, param, notFound string) (*T, error) {
	obj, err := orm.DeleteByID[T](c.DBContext(), c.Param(param))
	if err := NotFoundIf(err, notFound); err != nil {
		return nil, err
	}
	return obj, nil
}

// FetchQuery runs a custom queryset Get and maps sql.ErrNoRows to notFound.
func FetchQuery[T any](c *Context, query func(context.Context) (*T, error), notFound string) (*T, error) {
	obj, err := query(c.DBContext())
	if err := NotFoundIf(err, notFound); err != nil {
		return nil, err
	}
	return obj, nil
}

// ListQuery runs a custom queryset All.
func ListQuery[T any](c *Context, query func(context.Context) ([]*T, error)) ([]*T, error) {
	return query(c.DBContext())
}

// ListRespond loads a list and writes the standard success envelope.
func ListRespond[T any](c *Context, message string, query func(context.Context) ([]*T, error)) error {
	items, err := ListQuery[T](c, query)
	if err != nil {
		return err
	}
	if items == nil {
		items = make([]*T, 0)
	}
	return c.OK(message, items)
}

// ListRespondPaginated loads a paginated list and writes paginated envelope.
func ListRespondPaginated[T any](
	c *Context,
	message string,
	query func(context.Context, int, int) ([]*T, error),
	count func(context.Context) (int64, error),
) error {
	page, limit, offset := c.Pagination(20, 100)
	items, err := query(c.DBContext(), limit, offset)
	if err != nil {
		return err
	}
	if items == nil {
		items = make([]*T, 0)
	}
	total, err := count(c.DBContext())
	if err != nil {
		return err
	}
	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return c.Paginated(message, items, PageMeta{
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	})
}
