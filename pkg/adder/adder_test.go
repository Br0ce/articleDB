package adder

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/ids"
	"github.com/Br0ce/articleDB/pkg/logger"
	"github.com/Br0ce/articleDB/pkg/mock"
)

func TestAdder_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		log          *zap.SugaredLogger
		summarizerFn func(ctx context.Context, text string) (string, error)
		nerizerFn    func(ctx context.Context, text string) (article.NER, error)
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
				summarizerFn: func(ctx context.Context, txt string) (string, error) {
					if txt != body {
						t.Fatalf("summarizer text not equal, want %s got %s", body, txt)
					}
					return "Summary of text.", nil
				},
				nerizerFn: func(ctx context.Context, txt string) (article.NER, error) {
					if txt != body {
						t.Fatalf("ner text not equal, want %s got %s", body, txt)
					}
					return article.NER{}, nil
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
				summarizerFn: func(ctx context.Context, txt string) (string, error) {
					return "", errors.New("summarizer error")
				},
				nerizerFn: func(ctx context.Context, txt string) (article.NER, error) {
					return article.NER{}, nil
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
				summarizerFn: func(ctx context.Context, txt string) (string, error) {
					return "Summary of text.", nil
				},
				nerizerFn: func(ctx context.Context, txt string) (article.NER, error) {
					return article.NER{}, errors.New("ner error")
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
	}

	for _, tt := range tests {
		summarizer := &mock.Summarizer{SummarizeFn: tt.fields.summarizerFn}
		ner := &mock.NER{NERFn: tt.fields.nerizerFn}

		t.Run(tt.name, func(t *testing.T) {
			a := New(summarizer, ner, tt.fields.log)

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
