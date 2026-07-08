package gokil-cli-demo

import (
	"gokil-cli-demo/views"
	"github.com/lrndwy/gokil/framework"
	"github.com/lrndwy/gokil/router"
)

func URLPatterns(app *framework.App, r *router.Router) {
	r.GET("/api/health/", app.Wrap(views.HealthCheck))
	r.GET("/api/users/", app.Wrap(views.UserList))
	r.POST("/api/users/", app.Wrap(views.UserCreate))
	r.GET("/api/users/:id", app.Wrap(views.UserDetail))
	r.PUT("/api/users/:id", app.Wrap(views.UserUpdate))
	r.DELETE("/api/users/:id", app.Wrap(views.UserDelete))
	r.GET("/api/posts/", app.Wrap(views.PostList))
	r.POST("/api/posts/", app.Wrap(views.PostCreate))
	r.GET("/api/posts/:id", app.Wrap(views.PostDetail))
	r.PUT("/api/posts/:id", app.Wrap(views.PostUpdate))
	r.DELETE("/api/posts/:id", app.Wrap(views.PostDelete))
	r.GET("/api/tags/", app.Wrap(views.TagList))
	r.POST("/api/tags/", app.Wrap(views.TagCreate))
	r.GET("/api/tags/:id", app.Wrap(views.TagDetail))
	r.PUT("/api/tags/:id", app.Wrap(views.TagUpdate))
	r.DELETE("/api/tags/:id", app.Wrap(views.TagDelete))
}
