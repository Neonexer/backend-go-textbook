package repository

import (
	"time"

	"github.com/go-course/project-1-blog/internal/model"
)

// MemoryRepo — in-memory хранилище постов.
type MemoryRepo struct {
	posts  []model.Post
	nextID int
}

// NewMemory создаёт хранилище с начальными данными.
func NewMemory() *MemoryRepo {
	return &MemoryRepo{
		posts: []model.Post{
			{ID: 1, Title: "Первый пост", Body: "Привет, мир!", Status: model.StatusPublished, CreatedAt: time.Now()},
			{ID: 2, Title: "Второй пост", Body: "Изучаем Go и chi.", Status: model.StatusDraft, CreatedAt: time.Now()},
		},
		nextID: 3,
	}
}

func (r *MemoryRepo) FindAll() []model.Post {
	return r.posts
}

func (r *MemoryRepo) FindByID(id int) (model.Post, bool) {
	for _, p := range r.posts {
		if p.ID == id {
			return p, true
		}
	}
	return model.Post{}, false
}

func (r *MemoryRepo) Create(p model.Post) model.Post {
	p.ID = r.nextID
	r.nextID++
	r.posts = append(r.posts, p)
	return p
}

func (r *MemoryRepo) Update(id int, p model.Post) (model.Post, bool) {
	for i, existing := range r.posts {
		if existing.ID == id {
			p.ID = id
			p.CreatedAt = existing.CreatedAt
			r.posts[i] = p
			return p, true
		}
	}
	return model.Post{}, false
}

func (r *MemoryRepo) Delete(id int) bool {
	for i, p := range r.posts {
		if p.ID == id {
			r.posts = append(r.posts[:i], r.posts[i+1:]...)
			return true
		}
	}
	return false
}
