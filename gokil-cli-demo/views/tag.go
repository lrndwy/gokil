package views

import (
	"context"

	"gokil-cli-demo/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	return views.List(ctx, "tags retrieved", orm.Objects[models.Tag](ctx.DBContext()))
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
	return views.Create(ctx, "tag", &models.Tag{Name: input.Name})
}

func TagDetail(ctx *views.Context) error {
	return views.DetailByID[models.Tag](ctx, "id", "tag", "tag not found")
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
	return views.Update[models.Tag](ctx, "id", "tag", "tag not found", map[string]any{
		"name": input.Name,
	})
}

func TagDelete(ctx *views.Context) error {
	return views.Delete[models.Tag](ctx, "id", "tag", "tag not found")
}
