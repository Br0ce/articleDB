package mock

import "context"

type Summarizer struct {
	SummarizeFn       func(ctx context.Context, text string) (string, error)
	SummarizerInvoked bool
}

func (s *Summarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.SummarizerInvoked = true
	return s.SummarizeFn(ctx, text)
}

type NERer struct {
	NERFn      func(ctx context.Context, text string) ([]string, []string, []string, error)
	NERInvoked bool
}

func (n *NERer) NER(ctx context.Context, text string) ([]string, []string, []string, error) {
	n.NERInvoked = true
	return n.NERFn(ctx, text)
}
