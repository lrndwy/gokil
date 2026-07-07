package views

import (
	"demoapi/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func getPost(ctx *views.Context, id string) (*models.Post, error) {
	post, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		PrefetchRelated("Tags").
		Filter("id", id).
		Get()
	if err := views.NotFoundIf(err, "post not found"); err != nil {
		return nil, err
	}
	return post, nil
}

func PostList(ctx *views.Context) error {
	posts, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		All()
	if err != nil {
		return err
	}
	return ctx.OK("posts retrieved", posts)
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
	post, err := orm.Create(ctx.DBContext(), &models.Post{
		Title:    input.Title,
		Content:  input.Content,
		AuthorID: input.AuthorID,
	})
	if err != nil {
		return err
	}
	return ctx.Created("post created", post)
}

func PostDetail(ctx *views.Context) error {
	post, err := getPost(ctx, ctx.Param("id"))
	if err != nil {
		return err
	}
	return ctx.OK("post retrieved", post)
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
	_, err := orm.UpdateByID[models.Post](ctx.DBContext(), ctx.Param("id"), map[string]any{
		"title":     input.Title,
		"content":   input.Content,
		"author_id": input.AuthorID,
	})
	if err := views.NotFoundIf(err, "post not found"); err != nil {
		return err
	}
	post, err := getPost(ctx, ctx.Param("id"))
	if err != nil {
		return err
	}
	return ctx.OK("post updated", post)
}

func PostDelete(ctx *views.Context) error {
	post, err := orm.DeleteByID[models.Post](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "post not found"); err != nil {
		return err
	}
	return ctx.OK("post deleted", post)
}
