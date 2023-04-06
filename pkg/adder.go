package pkg

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type NERer interface {
	NER(ctx context.Context, text string) ([]string, []string, []string, error)
}

type ArticleAdder struct {
	ner        NERer
	summarizer Summarizer
	log        *zap.SugaredLogger
}

func NewArticleAdder(summarizer Summarizer, ner NERer, log *zap.SugaredLogger) *ArticleAdder {
	return &ArticleAdder{
		ner:        ner,
		summarizer: summarizer,
		log:        log}
}

func (a *ArticleAdder) Add(ctx context.Context, article Article) error {
	a.log.Infow("add article", "method", "Add", "articleID", article.ID)

	if !ValidID(article.ID) {
		return ErrInvalidID
	}

	g, ctx := errgroup.WithContext(ctx)

	a.log.Debugw("start requests", "method", "Add", "articleID", article.ID)
	g.Go(func() error {
		summary, err := a.summarizer.Summarize(ctx, article.Body)
		if err != nil {
			return err
		}
		article.Summary = summary
		return nil
	})

	g.Go(func() error {
		pers, locs, orgs, err := a.ner.NER(ctx, article.Body)
		if err != nil {
			return err
		}
		article.Pers = pers
		article.Locs = locs
		article.Orgs = orgs
		return nil
	})

	a.log.Debugw("wait for requests to finish", "method", "Add", "articleID", article.ID)
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
