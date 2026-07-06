package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupTestRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	setupRoutes(r)
	return r
}

func TestListPosts(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d", rec.Code)
	}

	var result []Post
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("ошибка декодирования: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("ожидалось минимум 2 поста, получено %d", len(result))
	}
}

func TestGetPost_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d", rec.Code)
	}
}

func TestGetPost_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/posts/999", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ожидался 404, получен %d", rec.Code)
	}
}

func TestCreatePost_Success(t *testing.T) {
	body := strings.NewReader(`{"title":"Новый","body":"Текст"}`)
	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Authorization", "Bearer test")
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался 201, получен %d: %s", rec.Code, rec.Body.String())
	}

	var p Post
	json.NewDecoder(rec.Body).Decode(&p)
	if p.Title != "Новый" {
		t.Errorf("ожидался title 'Новый', получен '%s'", p.Title)
	}
	if p.Status != StatusDraft {
		t.Errorf("ожидался статус draft по умолчанию")
	}
	if p.CreatedAt.IsZero() {
		t.Errorf("ожидался CreatedAt, получен zero value")
	}
}

func TestCreatePost_NoAuth(t *testing.T) {
	body := strings.NewReader(`{"title":"X","body":"Y"}`)
	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401, получен %d", rec.Code)
	}
}

func TestCreatePost_Validation(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"empty title", `{"title":"","body":"text"}`, http.StatusBadRequest},
		{"empty body", `{"title":"title","body":""}`, http.StatusBadRequest},
		{"title too long", `{"title":"` + strings.Repeat("x", 201) + `","body":"text"}`, http.StatusBadRequest},
		{"missing field", `{"title":"ok","body":"ok","extra_field":"?"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer test")
			rec := httptest.NewRecorder()

			setupTestRouter().ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Errorf("ожидался %d, получен %d: %s", tt.status, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUpdatePost_NotFound(t *testing.T) {
	body := strings.NewReader(`{"title":"Upd","body":"..."}`)
	req := httptest.NewRequest(http.MethodPut, "/posts/999", body)
	req.Header.Set("Authorization", "Bearer test")
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ожидался 404, получен %d", rec.Code)
	}
}

func TestDeletePost_NoAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401, получен %d", rec.Code)
	}
}
