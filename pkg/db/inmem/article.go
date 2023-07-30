package inmem

import (
	"context"
	"sync"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/db"
	"github.com/Br0ce/articleDB/pkg/ids"
)

// Article is an inmemory implemetation for the article.DB interface.
type Article struct {
	items map[string]article.Article
	mu    sync.RWMutex
}

// NewArticle is a factory for inmem.Article, that implements the
// article.DB interface.
func NewArticle() *Article {
	return &Article{
		items: make(map[string]article.Article),
	}
}

// Add adds an article.Article to the db and returns it assigend ID for retrieval.
// The ID will be assigned to the article.Article.ID field by overriding its old value.
func (a *Article) Add(ctx context.Context, item article.Article) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	id := ids.UniqueID()
	item.ID = id

	a.items[id] = item

	return id, nil
}

// Get returns the article.Article for the given ID.
func (a *Article) Get(ctx context.Context, id string) (article.Article, error) {
	if !ids.ValidID(id) {
		return article.Article{}, ids.ErrInvalidID
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	item, ok := a.items[id]
	if !ok {
		return article.Article{}, db.ErrNotFound
	}

	return item, nil
}
