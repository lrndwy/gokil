package views

import (
	"context"

	"github.com/lrndwy/gokil/orm"
)

// DetailByID loads one record by route param and writes "<resource> retrieved".
func DetailByID[T any](c *Context, param, resource, notFound string) error {
	obj, err := FetchByIDParam[T](c, param, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("retrieved", resource, obj)
}

// UpdateByParam updates one record by route param and writes "<resource> updated".
func UpdateByParam[T any](c *Context, param, resource, notFound string, values map[string]any) error {
	obj, err := UpdateByIDParam[T](c, param, values, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("updated", resource, obj)
}

// DeleteByParam deletes one record by route param and writes "<resource> deleted".
func DeleteByParam[T any](c *Context, param, resource, notFound string) error {
	obj, err := DeleteByIDParam[T](c, param, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("deleted", resource, obj)
}

// CreateAndRespond creates a record and writes "<resource> created".
func CreateAndRespond[T any](c *Context, resource string, create func(context.Context) (*T, error)) error {
	obj, err := create(c.DBContext())
	if err != nil {
		return err
	}
	return c.ResourceCreated(resource, obj)
}

// Create creates a record using orm.Create and writes "<resource> created".
//
// This helper exists to avoid boilerplate closures in handlers:
//
//	return views.Create(ctx, "user", &models.User{Email: input.Email, Name: input.Name})
func Create[T any](c *Context, resource string, obj *T) error {
	created, err := orm.Create(c.DBContext(), obj)
	if err != nil {
		return err
	}
	return c.ResourceCreated(resource, created)
}

// Update updates one record by route param and writes "<resource> updated".
//
// This is the most common update helper (no refresh):
//
//	return views.Update[models.User](ctx, "id", "user", "user not found", map[string]any{"name": input.Name})
func Update[T any](c *Context, param, resource, notFound string, values map[string]any) error {
	obj, err := UpdateByIDParam[T](c, param, values, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("updated", resource, obj)
}

// Delete deletes one record by route param and writes "<resource> deleted".
func Delete[T any](c *Context, param, resource, notFound string) error {
	obj, err := DeleteByIDParam[T](c, param, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("deleted", resource, obj)
}

// DetailByQuery loads one record using a custom queryset and writes "<resource> retrieved".
func DetailByQuery[T any](
	c *Context,
	resource, notFound string,
	query func(context.Context) (*T, error),
) error {
	obj, err := FetchQuery(c, query, notFound)
	if err != nil {
		return err
	}
	return c.ResourceOK("retrieved", resource, obj)
}

// UpdateAndRefresh updates by param then reloads with refresh query before responding.
func UpdateAndRefresh[T any](
	c *Context,
	param, resource, notFound string,
	values map[string]any,
	refresh func(context.Context, string) (*T, error),
) error {
	if _, err := UpdateByIDParam[T](c, param, values, notFound); err != nil {
		return err
	}
	obj, err := refresh(c.DBContext(), c.Param(param))
	if err := NotFoundIf(err, notFound); err != nil {
		return err
	}
	return c.ResourceOK("updated", resource, obj)
}
