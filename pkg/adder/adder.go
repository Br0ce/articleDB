package adder

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/ids"
)

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type NamedEntityRecognizer interface {
	NER(ctx context.Context, text string) (article.NER, error)
}

type Adder struct {
	nerizer    NamedEntityRecognizer
	summarizer Summarizer
	log        *zap.SugaredLogger
}

func New(summarizer Summarizer, nerizer NamedEntityRecognizer, log *zap.SugaredLogger) *Adder {
	return &Adder{
		nerizer:    nerizer,
		summarizer: summarizer,
		log:        log}
}

func (a *Adder) Add(ctx context.Context, art article.Article) error {
	a.log.Infow("add article", "method", "Add", "articleID", art.ID)

	if !ids.ValidID(art.ID) {
		return ids.ErrInvalidID
	}

	art, err := a.addFeatures(ctx, art)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adder) addFeatures(ctx context.Context, art article.Article) (article.Article, error) {
	a.log.Infow("extract features and add to article", "method", "addFeatures", "articleID", art.ID)
	g, ctx := errgroup.WithContext(ctx)

	a.log.Debugw("start extracting features ...", "method", "addFeatures", "articleID", art.ID)
	g.Go(func() error {
		summary, err := a.summarizer.Summarize(ctx, art.Body)
		if err != nil {
			return err
		}
		art.Summary = summary
		return nil
	})

	g.Go(func() error {
		ner, err := a.nerizer.NER(ctx, art.Body)
		if err != nil {
			return err
		}
		art.NER = ner
		return nil
	})

	a.log.Debugw("wait for feature extraction to finish", "method", "addFeatures", "articleID", art.ID)
	if err := g.Wait(); err != nil {
		return article.Article{}, err
	}

	return art, nil
}
