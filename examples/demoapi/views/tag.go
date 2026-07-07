package views

import (
	"demoapi/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	tags, err := orm.Objects[models.Tag](ctx.DBContext()).All()
	if err != nil {
		return err
	}
	return ctx.OK("tags retrieved", tags)
}

func TagCreate(ctx *views.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	tag, err := orm.Create(ctx.DBContext(), &models.Tag{Name: input.Name})
	if err != nil {
		return err
	}
	return ctx.Created("tag created", tag)
}

func TagDetail(ctx *views.Context) error {
	tag, err := orm.GetByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	return ctx.OK("tag retrieved", tag)
}

func TagUpdate(ctx *views.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	tag, err := orm.UpdateByID[models.Tag](ctx.DBContext(), ctx.Param("id"), map[string]any{
		"name": input.Name,
	})
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	return ctx.OK("tag updated", tag)
}

func TagDelete(ctx *views.Context) error {
	tag, err := orm.DeleteByID[models.Tag](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "tag not found"); err != nil {
		return err
	}
	return ctx.OK("tag deleted", tag)
}
