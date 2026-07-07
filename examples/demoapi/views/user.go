package views

import (
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
	return ctx.OK("users retrieved", users)
}

func UserCreate(ctx *views.Context) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	user, err := orm.Create(ctx.DBContext(), &models.User{
		Email: input.Email,
		Name:  input.Name,
	})
	if err != nil {
		return err
	}
	return ctx.Created("user created", user)
}

func UserDetail(ctx *views.Context) error {
	user, err := orm.GetByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	return ctx.OK("user retrieved", user)
}

func UserUpdate(ctx *views.Context) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := ctx.MustBindJSON(&input); err != nil {
		return err
	}
	user, err := orm.UpdateByID[models.User](ctx.DBContext(), ctx.Param("id"), map[string]any{
		"email": input.Email,
		"name":  input.Name,
	})
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	return ctx.OK("user updated", user)
}

func UserDelete(ctx *views.Context) error {
	user, err := orm.DeleteByID[models.User](ctx.DBContext(), ctx.Param("id"))
	if err := views.NotFoundIf(err, "user not found"); err != nil {
		return err
	}
	return ctx.OK("user deleted", user)
}
