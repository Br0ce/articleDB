package pkg

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type NamedEntityRecognizer interface {
	NER(ctx context.Context, text string) ([]string, []string, []string, error)
}

type ArticleAdder struct {
	nerizer    NamedEntityRecognizer
	summarizer Summarizer
	log        *zap.SugaredLogger
}

func NewArticleAdder(summarizer Summarizer, nerizer NamedEntityRecognizer, log *zap.SugaredLogger) *ArticleAdder {
	return &ArticleAdder{
		nerizer:    nerizer,
		summarizer: summarizer,
		log:        log}
}

func (a *ArticleAdder) Add(ctx context.Context, article Article) error {
	a.log.Infow("add article", "method", "Add", "articleID", article.ID)

	if !ValidID(article.ID) {
		return ErrInvalidID
	}

	article, err := a.addFeatures(ctx, article)
	if err != nil {
		return err
	}

	return nil
}

func (a *ArticleAdder) addFeatures(ctx context.Context, article Article) (Article, error) {
	a.log.Infow("extract features and add to article", "method", "addFeatures", "articleID", article.ID)
	g, ctx := errgroup.WithContext(ctx)

	a.log.Debugw("start extracting features ...", "method", "addFeatures", "articleID", article.ID)
	g.Go(func() error {
		summary, err := a.summarizer.Summarize(ctx, article.Body)
		if err != nil {
			return err
		}
		article.Summary = summary
		return nil
	})

	g.Go(func() error {
		pers, locs, orgs, err := a.nerizer.NER(ctx, article.Body)
		if err != nil {
			return err
		}
		article.Pers = pers
		article.Locs = locs
		article.Orgs = orgs
		return nil
	})

	a.log.Debugw("wait for feature extraction to finish", "method", "addFeatures", "articleID", article.ID)
	if err := g.Wait(); err != nil {
		return Article{}, err
	}

	return article, nil
}
