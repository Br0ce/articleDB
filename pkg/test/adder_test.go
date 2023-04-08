package test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	articleDB "github.com/Br0ce/articleDB/pkg"
	"github.com/Br0ce/articleDB/pkg/logger"
	"github.com/Br0ce/articleDB/pkg/mock"
)

func TestArticleAdder_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		log          *zap.SugaredLogger
		summarizerFn func(ctx context.Context, text string) (string, error)
		nerizerFn    func(ctx context.Context, text string) ([]string, []string, []string, error)
	}
	type args struct {
		ctx     context.Context
		article articleDB.Article
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
				article: articleDB.Article{ID: ""},
			},
			wantErr: true,
			errMsg:  articleDB.ErrInvalidID.Error(),
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
				nerizerFn: func(ctx context.Context, txt string) ([]string, []string, []string, error) {
					if txt != body {
						t.Fatalf("ner text not equal, want %s got %s", body, txt)
					}
					return []string{}, []string{}, []string{}, nil
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: articleDB.Article{
					ID:   articleDB.UniqueID(),
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
				nerizerFn: func(ctx context.Context, txt string) ([]string, []string, []string, error) {
					return []string{}, []string{}, []string{}, nil
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: articleDB.Article{
					ID: articleDB.UniqueID(),
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
				nerizerFn: func(ctx context.Context, txt string) ([]string, []string, []string, error) {
					return nil, nil, nil, errors.New("ner error")
				},
				log: log},
			args: args{
				ctx: context.TODO(),
				article: articleDB.Article{
					ID:   articleDB.UniqueID(),
					Body: body,
				},
			},
			wantErr: true,
			errMsg:  "ner error",
		},
	}

	for _, tt := range tests {
		summarizer := &mock.Summarizer{SummarizeFn: tt.fields.summarizerFn}
		ner := &mock.NERer{NERFn: tt.fields.nerizerFn}

		t.Run(tt.name, func(t *testing.T) {
			a := articleDB.NewArticleAdder(summarizer, ner, tt.fields.log)

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
