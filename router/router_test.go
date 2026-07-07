package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lrndwy/gokil/router"
)

func TestPathParams(t *testing.T) {
	r := router.New()
	r.GET("/api/users/:id", func(w http.ResponseWriter, _ *http.Request, params map[string]string) {
		_, _ = w.Write([]byte(params["id"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/42", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Body.String() != "42" {
		t.Fatalf("expected 42, got %s", rec.Body.String())
	}
}

func TestNotFound(t *testing.T) {
	r := router.New()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
