package views

import (
	"context"
	"net/http"

	"demoapi/models"
	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/views"
)

func HealthCheck(ctx *views.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func UserList(ctx *views.Context) error {
	return views.List(ctx, "users retrieved", orm.Objects[models.User](ctx.DBContext()))
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
	return views.Create(ctx, "user", &models.User{Email: input.Email, Name: input.Name})
}

func UserDetail(ctx *views.Context) error {
	return views.DetailByID[models.User](ctx, "id", "user", "user not found")
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
	return views.UpdateByParam[models.User](ctx, "id", "user", "user not found", map[string]any{
		"email": input.Email,
		"name":  input.Name,
	})
}

func UserDelete(ctx *views.Context) error {
	return views.DeleteByParam[models.User](ctx, "id", "user", "user not found")
}
