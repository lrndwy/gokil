package views

import (
	"net/http"

	"gokil-cli-demo/models"
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
	return views.Listed(ctx, users, "users retrieved")
}

func UserCreate(ctx *views.Context) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"email": input.Email,
		"name":  input.Name,
	}); err != nil {
		return err
	}
	created, err := orm.Create(ctx.DBContext(), &models.User{Email: input.Email, Name: input.Name})
	if err != nil {
		return err
	}
	return views.Created(ctx, created, "users created")
}

func UserDetail(ctx *views.Context) error {
	user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	return views.Detailed(ctx, user, "user retrieved")
}

func UserUpdate(ctx *views.Context) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	if err := views.RequiredFields(map[string]string{
		"email": input.Email,
		"name":  input.Name,
	}); err != nil {
		return err
	}
	_, err := orm.UpdateByID[models.User](ctx.DBContext(), ctx.Param("id"), map[string]any{
		"email": input.Email,
		"name":  input.Name,
	})
	if err != nil {
		return err
	}
	userUpdated, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	return views.Updated(ctx, userUpdated, "users updated")
}

func UserDelete(ctx *views.Context) error {
	user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	_, err = orm.DeleteByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err != nil {
		return err
	}
	return views.Deleted(ctx, user, "users deleted")
}
