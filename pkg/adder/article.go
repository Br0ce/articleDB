package adder

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/vector"
)

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type NamedEntityRecognizer interface {
	NER(ctx context.Context, text string) (article.NER, error)
}

type Encoder interface {
	Encode(ctx context.Context, texts []string) ([]vector.Vector, error)
}

type Adder struct {
	ner     NamedEntityRecognizer
	sum     Summarizer
	encoder Encoder
	db      article.DB
	log     *slog.Logger
}

type AdderOption func(a *Adder)

func New(opts ...AdderOption) (*Adder, error) {
	adder := &Adder{}

	for _, opt := range opts {
		opt(adder)
	}

	if adder.sum == nil {
		return nil, errors.New("summarizer is nil")
	}

	if adder.ner == nil {
		return nil, errors.New("nerer is nil")
	}

	if adder.encoder == nil {
		return nil, errors.New("encoder is nil")
	}

	if adder.log == nil {
		return nil, errors.New("logger is nil")
	}

	return adder, nil
}

func WithSummarizer(sum Summarizer) AdderOption {
	return func(a *Adder) {
		a.sum = sum
	}
}

func WithNamedEntityRecognizer(ner NamedEntityRecognizer) AdderOption {
	return func(a *Adder) {
		a.ner = ner
	}
}

func WithEncoder(encoder Encoder) AdderOption {
	return func(a *Adder) {
		a.encoder = encoder
	}
}

func WithDB(db article.DB) AdderOption {
	return func(a *Adder) {
		a.db = db
	}
}

func WithLogger(log *slog.Logger) AdderOption {
	return func(a *Adder) {
		a.log = log
	}
}

func (a *Adder) Add(ctx context.Context, ar article.Article) (string, error) {
	a.log.Info("add article", "method", "Add", "articleID", ar.ID)

	ar, err := a.addFeatures(ctx, ar)
	if err != nil {
		return "", err
	}

	id, err := a.db.Add(ctx, ar)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (a *Adder) addFeatures(ctx context.Context, ar article.Article) (article.Article, error) {
	a.log.Info("extract features and add to article", "method", "addFeatures", "articleID", ar.ID)
	g, ctx := errgroup.WithContext(ctx)

	a.log.Debug("start extracting features ...", "method", "addFeatures", "articleID", ar.ID)
	g.Go(func() error {
		sum, err := a.sum.Summarize(ctx, ar.Body)
		if err != nil {
			return err
		}
		ar.Summary = sum
		return nil
	})

	g.Go(func() error {
		ner, err := a.ner.NER(ctx, ar.Body)
		if err != nil {
			return err
		}
		ar.NER = ner
		return nil
	})

	g.Go(func() error {
		_, err := a.encoder.Encode(ctx, []string{ar.Body})
		if err != nil {
			return err
		}
		return nil
	})

	a.log.Debug("wait for feature extraction to finish", "method", "addFeatures", "articleID", ar.ID)
	if err := g.Wait(); err != nil {
		return article.Article{}, err
	}

	return ar, nil
}
