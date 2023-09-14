package adder

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/extract/noop"
	"github.com/Br0ce/articleDB/pkg/logger"
	"github.com/Br0ce/articleDB/pkg/mock"
)

func TestAdder_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		log   *slog.Logger
		sumFn func(ctx context.Context, text string) (string, error)
		nerFn func(ctx context.Context, text string) (article.NER, error)
		addFn func(ctx context.Context, ar article.Article) (string, error)
	}

	type args struct {
		ctx     context.Context
		article article.Article
	}

	log := logger.NewTest(true)
	body := "This is a test body."

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
		errMsg  string
	}{
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
					return "1234", nil
				},
				log: log,
			},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					Body: body,
				},
			},
			wantErr: false,
			want:    "1234",
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
				log: log,
			},
			args: args{
				ctx:     context.TODO(),
				article: article.Article{},
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
				log: log,
			},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
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
				log: log,
			},
			args: args{
				ctx: context.TODO(),
				article: article.Article{
					Body: body,
				},
			},
			wantErr: true,
			errMsg:  "db error",
		},
	}

	for _, tt := range tests {
		sun := &mock.Summarizer{SummarizeFn: tt.fields.sumFn}
		ner := &mock.NER{NERFn: tt.fields.nerFn}
		db := &mock.DB{AddFn: tt.fields.addFn}

		t.Run(tt.name, func(t *testing.T) {
			a := Adder{
				sum: sun,
				ner: ner,
				db:  db,
				log: tt.fields.log}

			got, err := a.Add(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleAdder.Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && (tt.errMsg != err.Error()) {
				t.Errorf("errMsg want %v, got %v", tt.errMsg, err.Error())
			}

			if tt.wantErr {
				return
			}

			if tt.want != got {
				t.Errorf("ArticleAdder.Add() want = %s got %s", tt.want, got)
			}
		})
	}
}

func TestNewWith(t *testing.T) {
	t.Parallel()

	log := logger.NewTest(false)
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
				sum: noop,
				ner: noop,
				log: log,
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
