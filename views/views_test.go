package views_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lrndwy/gokil/views"
)

func TestHTTPError(t *testing.T) {
	err := views.NotFound("user not found")
	var httpErr *views.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatal("expected HTTPError")
	}
	if httpErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", httpErr.Status, http.StatusNotFound)
	}
}

func TestHandleErrorHTTPError(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &views.Context{Writer: rec}

	if err := views.HandleError(ctx, views.BadRequest("invalid json")); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "invalid json" {
		t.Fatalf("error = %v", body["error"])
	}
}

func TestHandleErrorGeneric(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &views.Context{Writer: rec}

	if err := views.HandleError(ctx, errors.New("db down")); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestNotFoundIf(t *testing.T) {
	if err := views.NotFoundIf(nil, "missing"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	var httpErr *views.HTTPError
	if err := views.NotFoundIf(sql.ErrNoRows, "missing"); !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if httpErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", httpErr.Status, http.StatusNotFound)
	}

	generic := errors.New("db down")
	if err := views.NotFoundIf(generic, "missing"); !errors.Is(err, generic) {
		t.Fatalf("expected generic error, got %v", err)
	}
}

func TestResponseEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &views.Context{Writer: rec}

	if err := ctx.OK("users retrieved", []string{"a"}); err != nil {
		t.Fatal(err)
	}

	var body views.Response
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != http.StatusOK {
		t.Fatalf("status = %d", body.Status)
	}
	if body.Message != "users retrieved" {
		t.Fatalf("message = %q", body.Message)
	}
}
