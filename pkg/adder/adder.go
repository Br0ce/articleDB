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
	nerer NamedEntityRecognizer
	sumer Summarizer
	db    article.DB
	log   *zap.SugaredLogger
}

func New(sumer Summarizer, nerer NamedEntityRecognizer, db article.DB, log *zap.SugaredLogger) *Adder {
	return &Adder{
		nerer: nerer,
		sumer: sumer,
		db:    db,
		log:   log}
}

func (a *Adder) Add(ctx context.Context, ar article.Article) error {
	a.log.Infow("add article", "method", "Add", "articleID", ar.ID)

	if !ids.ValidID(ar.ID) {
		return ids.ErrInvalidID
	}

	ar, err := a.addFeatures(ctx, ar)
	if err != nil {
		return err
	}

	_, err = a.db.Add(ctx, ar)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adder) addFeatures(ctx context.Context, ar article.Article) (article.Article, error) {
	a.log.Infow("extract features and add to article", "method", "addFeatures", "articleID", ar.ID)
	g, ctx := errgroup.WithContext(ctx)

	a.log.Debugw("start extracting features ...", "method", "addFeatures", "articleID", ar.ID)
	g.Go(func() error {
		sum, err := a.sumer.Summarize(ctx, ar.Body)
		if err != nil {
			return err
		}
		ar.Summary = sum
		return nil
	})

	g.Go(func() error {
		ner, err := a.nerer.NER(ctx, ar.Body)
		if err != nil {
			return err
		}
		ar.NER = ner
		return nil
	})

	a.log.Debugw("wait for feature extraction to finish", "method", "addFeatures", "articleID", ar.ID)
	if err := g.Wait(); err != nil {
		return article.Article{}, err
	}

	return ar, nil
}
