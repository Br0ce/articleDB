package adder

import (
	"context"
	"errors"

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

type AdderOption func(a *Adder)

func New(opts ...AdderOption) (*Adder, error) {
	adder := &Adder{}

	for _, opt := range opts {
		opt(adder)
	}

	if adder.sumer == nil {
		return nil, errors.New("summarizer is nil")
	}

	if adder.nerer == nil {
		return nil, errors.New("nerer is nil")
	}

	if adder.log == nil {
		return nil, errors.New("logger is nil")
	}

	return adder, nil
}

func WithSummarizer(sumer Summarizer) AdderOption {
	return func(a *Adder) {
		a.sumer = sumer
	}
}

func WithNamedEntityRecognizer(nerer NamedEntityRecognizer) AdderOption {
	return func(a *Adder) {
		a.nerer = nerer
	}
}

func WithDB(db article.DB) AdderOption {
	return func(a *Adder) {
		a.db = db
	}
}

func WithLogger(log *zap.SugaredLogger) AdderOption {
	return func(a *Adder) {
		a.log = log
	}
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
