package views

import (
	"gokil-cli-demo/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	tags, err := orm.Objects[models.Tag](ctx.DBContext()).All()
	if err != nil {
		return err
	}
	return views.Listed(ctx, tags, "tags retrieved")
}

func TagCreate(ctx *views.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.Required("name", input.Name); err != nil {
		return err
	}
	created, err := orm.Create(ctx.DBContext(), &models.Tag{Name: input.Name})
	if err != nil {
		return err
	}
	return views.Created(ctx, created, "tags created")
}

func TagDetail(ctx *views.Context) error {
	tag, err := orm.GetByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	return views.Detailed(ctx, tag, "tag retrieved")
}

func TagUpdate(ctx *views.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.Required("name", input.Name); err != nil {
		return err
	}
	_, err := orm.UpdateByID[models.Tag](ctx.DBContext(), ctx.Param("id"), map[string]any{
		"name": input.Name,
	})
	if err != nil {
		return err
	}
	tagUpdated, err := orm.GetByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	return views.Updated(ctx, tagUpdated, "tags updated")
}

func TagDelete(ctx *views.Context) error {
	tag, err := orm.GetByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	_, err = orm.DeleteByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err != nil {
		return err
	}
	return views.Deleted(ctx, tag, "tags deleted")
}
