package noop

import (
	"context"

	"github.com/Br0ce/articleDB/pkg/article"
)

type Client struct {
}

func (c Client) Summarize(ctx context.Context, text string) (string, error) {
	return "", nil
}

func (c Client) NER(ctx context.Context, text string) (article.NER, error) {
	return article.NER{}, nil
}
