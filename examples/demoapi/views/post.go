package views

import (
	"database/sql"
	"net/http"

	"demoapi/models"
	"gokil/orm"
	"gokil/views"
)

func PostList(ctx *views.Context) error {
	posts, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		All()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, posts)
}

func PostCreate(ctx *views.Context) error {
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID int64  `json:"author_id"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return views.Error(ctx, http.StatusBadRequest, "invalid json")
	}
	post, err := orm.Create(ctx.DBContext(), &models.Post{
		Title:    input.Title,
		Content:  input.Content,
		AuthorID: input.AuthorID,
	})
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, post)
}

func PostDetail(ctx *views.Context) error {
	post, err := orm.Objects[models.Post](ctx.DBContext()).
		SelectRelated("Author").
		PrefetchRelated("Tags").
		Filter("id", ctx.Param("id")).
		Get()
	if err == sql.ErrNoRows {
		return views.Error(ctx, http.StatusNotFound, "post not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, post)
}
