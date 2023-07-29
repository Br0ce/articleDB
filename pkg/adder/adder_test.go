package adder

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/client/noop"
	"github.com/Br0ce/articleDB/pkg/ids"
	"github.com/Br0ce/articleDB/pkg/logger"
	"github.com/Br0ce/articleDB/pkg/mock"
	"go.uber.org/zap"
)

func TestAdder_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		log   *zap.SugaredLogger
		sumFn func(ctx context.Context, text string) (string, error)
		nerFn func(ctx context.Context, text string) (article.NER, error)
		addFn func(ctx context.Context, ar article.Article) (string, error)
	}

	type args struct {
		ctx     context.Context
		article article.Article
	}

	log, err := logger.NewTest(true)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	body := "This is a test body."

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name:   "invalid id",
			fields: fields{log: log},
			args: args{
				ctx:     context.TODO(),
				article: article.Article{ID: ""},
			},
			wantErr: true,
			errMsg:  ids.ErrInvalidID.Error(),
		},
		{
			name: "pass",
			fields: fields{
				sumFn: func(ctx context.Context, txt string) (string, error) {
					if txt != body {
						t.Fatalf("summarizer text not equal, want %s got %s", body, txt)
					}
					return "Summary of text.", nil
				},
				nerFn: func(ctx context.Context, txt string) (article.NER, error) {
					if txt != body {
						t.Fatalf("ner text not equal, want %s got %s", body, txt)
					}
					return article.NER{}, nil
				},
				addFn: func(ctx context.Context, ar article.Article) (string, error) {
					return "", nil
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					ID:   ids.UniqueID(),
					Body: body,
				},
			},
			wantErr: false,
		},
		{
			name: "summarizer error",
			fields: fields{
				sumFn: func(ctx context.Context, txt string) (string, error) {
					return "", errors.New("summarizer error")
				},
				nerFn: func(ctx context.Context, txt string) (article.NER, error) {
					return article.NER{}, nil
				},
				addFn: func(ctx context.Context, ar article.Article) (string, error) {
					return "", nil
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					ID: ids.UniqueID(),
				},
			},
			wantErr: true,
			errMsg:  "summarizer error",
		},
		{
			name: "ner error",
			fields: fields{
				sumFn: func(ctx context.Context, txt string) (string, error) {
					return "Summary of text.", nil
				},
				nerFn: func(ctx context.Context, txt string) (article.NER, error) {
					return article.NER{}, errors.New("ner error")
				},
				addFn: func(ctx context.Context, ar article.Article) (string, error) {
					return "", nil
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					ID:   ids.UniqueID(),
					Body: body,
				},
			},
			wantErr: true,
			errMsg:  "ner error",
		},
		{
			name: "db error",
			fields: fields{
				sumFn: func(ctx context.Context, txt string) (string, error) {
					return "Summary of text.", nil
				},
				nerFn: func(ctx context.Context, txt string) (article.NER, error) {
					return article.NER{}, nil
				},
				addFn: func(ctx context.Context, ar article.Article) (string, error) {
					return "", errors.New("db error")
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					ID:   ids.UniqueID(),
					Body: body,
				},
			},
			wantErr: true,
			errMsg:  "db error",
		},
	}

	for _, tt := range tests {
		sumer := &mock.Summarizer{SummarizeFn: tt.fields.sumFn}
		nerer := &mock.NER{NERFn: tt.fields.nerFn}
		db := &mock.DB{AddFn: tt.fields.addFn}

		t.Run(tt.name, func(t *testing.T) {
			a := Adder{
				sumer: sumer,
				nerer: nerer,
				db:    db,
				log:   tt.fields.log}

			err := a.Add(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleAdder.Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && (tt.errMsg != err.Error()) {
				t.Errorf("errMsg want %v, got %v", tt.errMsg, err.Error())
			}

		})
	}
}

func TestNewWith(t *testing.T) {
	t.Parallel()

	log, err := logger.NewTest(false)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	noop := noop.Client{}

	tests := []struct {
		opts    []AdderOption
		name    string
		want    *Adder
		wantErr bool
	}{
		{
			name: "pass",
			opts: []AdderOption{
				WithSummarizer(noop),
				WithNamedEntityRecognizer(noop),
				WithLogger(log),
			},
			wantErr: false,
			want: &Adder{
				sumer: noop,
				nerer: noop,
				log:   log,
			},
		},
		{
			name: "no logger",
			opts: []AdderOption{
				WithNamedEntityRecognizer(noop),
				WithSummarizer(noop),
			},
			wantErr: true,
		},
		{
			name: "no summarizer",
			opts: []AdderOption{
				WithNamedEntityRecognizer(noop),
				WithLogger(log),
			},
			wantErr: true,
		},
		{
			name: "no named entity recognizer",
			opts: []AdderOption{
				WithSummarizer(noop),
				WithLogger(log),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewWith() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
