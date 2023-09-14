package mock

import (
	"context"

	"github.com/Br0ce/articleDB/pkg/article"
)

type Summarizer struct {
	SummarizeFn       func(ctx context.Context, text string) (string, error)
	SummarizerInvoked bool
}

func (s *Summarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.SummarizerInvoked = true
	return s.SummarizeFn(ctx, text)
}

type NER struct {
	NERFn      func(ctx context.Context, text string) (article.NER, error)
	NERInvoked bool
}

func (n *NER) NER(ctx context.Context, text string) (article.NER, error) {
	n.NERInvoked = true
	return n.NERFn(ctx, text)
}
