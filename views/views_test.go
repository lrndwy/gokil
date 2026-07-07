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

func TestRequiredFields(t *testing.T) {
	err := views.RequiredFields(map[string]string{
		"email": "",
		"name":  "Ali",
	})
	var httpErr *views.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if httpErr.Status != http.StatusBadRequest {
		t.Fatalf("status = %d", httpErr.Status)
	}
}

func TestResourceOK(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &views.Context{Writer: rec}

	if err := ctx.ResourceOK("retrieved", "user", map[string]string{"name": "Ali"}); err != nil {
		t.Fatal(err)
	}
	var body views.Response
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Message != "user retrieved" {
		t.Fatalf("message = %q", body.Message)
	}
}

func TestPagination(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=2&limit=10", nil)
	ctx := &views.Context{Request: req}

	page, limit, offset := ctx.Pagination(20, 100)
	if page != 2 || limit != 10 || offset != 10 {
		t.Fatalf("page=%d limit=%d offset=%d", page, limit, offset)
	}
}

func TestPaginatedResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &views.Context{Writer: rec}

	if err := ctx.Paginated("users retrieved", []string{"a"}, views.PageMeta{
		Total: 42,
		Page:  1,
		Limit: 20,
		Pages: 3,
	}); err != nil {
		t.Fatal(err)
	}

	var body views.PaginatedResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Meta.Total != 42 {
		t.Fatalf("total = %d", body.Meta.Total)
	}
}
