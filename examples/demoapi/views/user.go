package views

import (
	"database/sql"
	"net/http"

	"demoapi/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func HealthCheck(ctx *views.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func UserList(ctx *views.Context) error {
	users, err := orm.Objects[models.User](ctx.DBContext()).All()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, users)
}

func UserCreate(ctx *views.Context) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return views.Error(ctx, http.StatusBadRequest, "invalid json")
	}
	user, err := orm.Create(ctx.DBContext(), &models.User{
		Email: input.Email,
		Name:  input.Name,
	})
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, user)
}

func UserDetail(ctx *views.Context) error {
	user, err := orm.Objects[models.User](ctx.DBContext()).
		Filter("id", ctx.Param("id")).
		Get()
	if err == sql.ErrNoRows {
		return views.Error(ctx, http.StatusNotFound, "user not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, user)
}
