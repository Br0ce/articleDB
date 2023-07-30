package mock

import (
	"context"

	"github.com/Br0ce/articleDB/pkg/article"
)

type DB struct {
	AddFn      func(ctx context.Context, ar article.Article) (string, error)
	AddInvoked bool

	GetFn      func(ctx context.Context, id string) (article.Article, error)
	GetInvoked bool
}

func (db *DB) Add(ctx context.Context, ar article.Article) (string, error) {
	db.AddInvoked = true
	return db.AddFn(ctx, ar)
}

func (db *DB) Get(ctx context.Context, id string) (article.Article, error) {
	db.GetInvoked = true
	return db.GetFn(ctx, id)
}
