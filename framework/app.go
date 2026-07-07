package framework

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"gokil/config"
	"gokil/orm"
	"gokil/router"
	"gokil/storage"
	"gokil/views"
)

type URLConfigurator func(*router.Router)

type App struct {
	Settings config.Settings
	Router   *router.Router
	Storage  storage.Provider
	DB       *orm.DB
	server   *http.Server
}

func New(settings config.Settings, configureURLs URLConfigurator) (*App, error) {
	var db *orm.DB
	if settings.Database.DSN != "" {
		var err error
		db, err = orm.Connect(
			settings.Database.Driver,
			settings.Database.DSN,
			settings.Database.MaxOpenConns,
			settings.Database.MaxIdleConns,
		)
		if err != nil {
			return nil, fmt.Errorf("connect database: %w", err)
		}
	}

	provider, err := storage.New(settings.Storage)
	if err != nil {
		if db != nil {
			_ = db.Close()
		}
		return nil, err
	}

	r := router.New()
	r.GET("/healthz", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	if configureURLs != nil {
		configureURLs(r)
	}

	app := &App{
		Settings: settings,
		Router:   r,
		Storage:  provider,
		DB:       db,
	}

	r.Use(app.requestMiddleware())
	return app, nil
}

func (a *App) requestMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func (a *App) Wrap(handler views.Handler) router.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx := &views.Context{
			Request: r,
			Writer:  w,
			DB:      a.DB,
			Storage: a.Storage,
			Params:  params,
		}
		reqCtx := r.Context()
		if a.DB != nil {
			reqCtx = orm.WithDB(reqCtx, a.DB)
		}
		ctx.Request = r.WithContext(reqCtx)

		if err := handler(ctx); err != nil {
			_ = views.Error(ctx, http.StatusInternalServerError, err.Error())
		}
	}
}

func (a *App) Run(ctx context.Context) error {
	address := net.JoinHostPort(a.Settings.Host, strconv.Itoa(a.Settings.Port))
	a.server = &http.Server{
		Addr:    address,
		Handler: a.Router.Handler(),
	}

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx := context.Background()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if a.DB != nil {
			_ = a.DB.Close()
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("run server: %w", err)
		}
		return nil
	}
}

func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
