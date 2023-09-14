package article

import "context"

type DB interface {
	Add(ctx context.Context, ar Article) (string, error)
	Get(ctx context.Context, id string) (Article, error)
}
