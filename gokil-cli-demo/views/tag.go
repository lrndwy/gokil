package views

import (
	"context"

	"gokil-cli-demo/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func TagList(ctx *views.Context) error {
	return views.ListRespond(ctx, "tags retrieved", func(db context.Context) ([]*models.Tag, error) {
		return orm.Objects[models.Tag](db).All()
	})
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
	return views.CreateAndRespond(ctx, "tag", func(db context.Context) (*models.Tag, error) {
		return orm.Create(db, &models.Tag{Name: input.Name})
	})
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
	return views.UpdateByParam[models.Tag](ctx, "id", "tag", "tag not found", map[string]any{
		"name": input.Name,
	})
}

func TagDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.Tag](ctx, "id", "tag", "tag not found")
}
