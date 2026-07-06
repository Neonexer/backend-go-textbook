package service_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/go-course/project-1-blog/internal/model"
	"github.com/go-course/project-1-blog/internal/service"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

type mockRepo struct {
	posts []model.Post
}

func (m *mockRepo) FindAll() []model.Post { return m.posts }
func (m *mockRepo) FindByID(id int) (model.Post, bool) {
	for _, p := range m.posts {
		if p.ID == id {
			return p, true
		}
	}
	return model.Post{}, false
}
func (m *mockRepo) Create(p model.Post) model.Post { m.posts = append(m.posts, p); return p }
func (m *mockRepo) Update(id int, p model.Post) (model.Post, bool) {
	for i, ex := range m.posts {
		if ex.ID == id {
			m.posts[i] = p
			return p, true
		}
	}
	return model.Post{}, false
}
func (m *mockRepo) Delete(id int) bool {
	for i, p := range m.posts {
		if p.ID == id {
			m.posts = append(m.posts[:i], m.posts[i+1:]...)
			return true
		}
	}
	return false
}

func TestList(t *testing.T) {
	repo := &mockRepo{posts: []model.Post{{ID: 1, Title: "Test"}}}
	svc := service.NewPost(repo, testLogger)
	posts := svc.List()
	if len(posts) != 1 {
		t.Errorf("expected 1, got %d", len(posts))
	}
}

func TestGet_Found(t *testing.T) {
	repo := &mockRepo{posts: []model.Post{{ID: 1, Title: "Test"}}}
	svc := service.NewPost(repo, testLogger)
	p, err := svc.Get(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Title != "Test" {
		t.Errorf("expected 'Test', got '%s'", p.Title)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := service.NewPost(&mockRepo{}, testLogger)
	_, err := svc.Get(999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCreate_Success(t *testing.T) {
	svc := service.NewPost(&mockRepo{}, testLogger)
	p, err := svc.Create("New", "Body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Title != "New" || p.Status != model.StatusDraft {
		t.Errorf("unexpected post: %+v", p)
	}
}

func TestCreate_Validation(t *testing.T) {
	svc := service.NewPost(&mockRepo{}, testLogger)
	_, err := svc.Create("", "body")
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := service.NewPost(&mockRepo{}, testLogger)
	err := svc.Delete(999)
	if err == nil {
		t.Error("expected error for non-existent post")
	}
}
