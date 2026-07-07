package demoapi

import (
	"demoapi/views"
	"github.com/lrndwy/gokil/framework"
	"github.com/lrndwy/gokil/router"
)

func URLPatterns(app *framework.App, r *router.Router) {
	r.GET("/api/health/", app.Wrap(views.HealthCheck))
	r.GET("/api/users/", app.Wrap(views.UserList))
	r.POST("/api/users/", app.Wrap(views.UserCreate))
	r.GET("/api/users/:id", app.Wrap(views.UserDetail))
	r.GET("/api/posts/", app.Wrap(views.PostList))
	r.POST("/api/posts/", app.Wrap(views.PostCreate))
	r.GET("/api/posts/:id", app.Wrap(views.PostDetail))
}
