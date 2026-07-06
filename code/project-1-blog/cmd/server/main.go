package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// --- Модели ---

// Status — статус поста.
type Status int

const (
	StatusDraft     Status = 0
	StatusPublished Status = 1
)

func (s Status) MarshalJSON() ([]byte, error) {
	switch s {
	case StatusDraft:
		return json.Marshal("draft")
	case StatusPublished:
		return json.Marshal("published")
	default:
		return json.Marshal("unknown")
	}
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "draft":
		*s = StatusDraft
	case "published":
		*s = StatusPublished
	default:
		*s = StatusDraft
	}
	return nil
}

// Post — модель поста блога.
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate проверяет бизнес-правила.
func (p *Post) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len(p.Title) > 200 {
		return fmt.Errorf("title too long: %d chars (max 200)", len(p.Title))
	}
	if strings.TrimSpace(p.Body) == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// ErrorResponse — структура ошибки API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// in-memory хранилище.
var posts = []Post{
	{ID: 1, Title: "Первый пост", Body: "Привет, мир!", Status: StatusPublished, CreatedAt: time.Now()},
	{ID: 2, Title: "Второй пост", Body: "Изучаем Go и chi.", Status: StatusDraft, CreatedAt: time.Now()},
}

// --- Хелперы ответа ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func decodeJSON(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// --- Главная ---

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	setupRoutes(r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("Сервер запущен на http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}

func setupRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Get("/posts", listPosts)
		r.Get("/posts/{id:[0-9]+}", getPost)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/posts", createPost)
		r.Put("/posts/{id:[0-9]+}", updatePost)
		r.Delete("/posts/{id:[0-9]+}", deletePost)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Обработчики ---

func listPosts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, posts)
}

func getPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	for _, p := range posts {
		if fmt.Sprintf("%d", p.ID) == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeError(w, http.StatusNotFound, "post not found")
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var p Post
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := p.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	p.ID = len(posts) + 1
	p.CreatedAt = time.Now()
	if p.Status == 0 {
		p.Status = StatusDraft
	}
	posts = append(posts, p)
	writeJSON(w, http.StatusCreated, p)
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var updated Post
	if err := decodeJSON(r, &updated); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := updated.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	for i, p := range posts {
		if fmt.Sprintf("%d", p.ID) == id {
			updated.ID = p.ID
			updated.CreatedAt = p.CreatedAt
			posts[i] = updated
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}
	writeError(w, http.StatusNotFound, "post not found")
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	for i, p := range posts {
		if fmt.Sprintf("%d", p.ID) == id {
			posts = append(posts[:i], posts[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeError(w, http.StatusNotFound, "post not found")
}
