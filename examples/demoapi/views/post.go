package views

import (
	"context"

	"demoapi/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func getPost(ctx *views.Context, id string) (*models.Post, error) {
	return views.FetchQuery(ctx, func(db context.Context) (*models.Post, error) {
		return orm.Objects[models.Post](db).
			SelectRelated("Author").
			PrefetchRelated("Tags").
			Filter("id", id).
			Get()
	}, "post not found")
}

func PostList(ctx *views.Context) error {
	return views.ListRespond(ctx, "posts retrieved", func(db context.Context) ([]*models.Post, error) {
		return orm.Objects[models.Post](db).SelectRelated("Author").All()
	})
}

func PostCreate(ctx *views.Context) error {
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID int64  `json:"author_id"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	return views.CreateAndRespond(ctx, "post", func(db context.Context) (*models.Post, error) {
		return orm.Create(db, &models.Post{
			Title:    input.Title,
			Content:  input.Content,
			AuthorID: input.AuthorID,
		})
	})
}

func PostDetail(ctx *views.Context) error {
	return views.DetailByQuery(ctx, "post", "post not found", func(db context.Context) (*models.Post, error) {
		return orm.Objects[models.Post](db).
			SelectRelated("Author").
			PrefetchRelated("Tags").
			Filter("id", ctx.Param("id")).
			Get()
	})
}

func PostUpdate(ctx *views.Context) error {
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID int64  `json:"author_id"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"title": input.Title,
	}); err != nil {
		return err
	}
	return views.UpdateAndRefresh(ctx, "id", "post", "post not found", map[string]any{
		"title":     input.Title,
		"content":   input.Content,
		"author_id": input.AuthorID,
	}, func(db context.Context, id string) (*models.Post, error) {
		return getPost(ctx, id)
	})
}

func PostDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.Post](ctx, "id", "post", "post not found")
}
