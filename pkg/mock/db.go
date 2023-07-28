package mock

import (
	"context"

	"github.com/Br0ce/articleDB/pkg/article"
)

type DB struct {
	AddFn      func(ctx context.Context, ar article.Article) (string, error)
	AddInvoked bool
}

func (db *DB) Add(ctx context.Context, ar article.Article) (string, error) {
	db.AddInvoked = true
	return db.AddFn(ctx, ar)
}
