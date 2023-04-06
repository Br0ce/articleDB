package test

import (
	"context"
	"net/url"
	"testing"
	"time"

	articleDB "github.com/Br0ce/articleDB/pkg"
	"github.com/Br0ce/articleDB/pkg/logger"
	"go.uber.org/zap"
)

func TestArticle_Equal(t *testing.T) {
	t.Parallel()

	type fields struct {
		Title     string
		Addr      url.URL
		Author    string
		Published time.Time
		Body      string
	}
	type args struct {
		b articleDB.Article
	}

	title := "Test Title"
	addr := url.URL{
		Scheme: "https",
		Host:   "testNewsroom",
		Path:   "tests/test",
	}
	author := "John Doe"
	puplished := time.Now().UTC()
	body := "Some article body."

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "pass",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     title,
				Addr:      addr,
				Author:    author,
				Published: puplished,
				Body:      body,
			}},
			want: true,
		},
		{
			name:   "different title",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     "Some Other Title",
				Addr:      addr,
				Author:    author,
				Published: puplished,
				Body:      body,
			}},
			want: false,
		},
		{
			name:   "different addr",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     title,
				Addr:      *addr.JoinPath("different"),
				Author:    author,
				Published: puplished,
				Body:      body,
			}},
			want: false,
		},
		{
			name:   "different published date",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     title,
				Addr:      addr,
				Author:    author,
				Published: puplished.Add(time.Hour),
				Body:      body,
			}},
			want: false,
		},
		{
			name:   "different body",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     title,
				Addr:      addr,
				Author:    author,
				Published: puplished,
				Body:      "Some other body.",
			}},
			want: false,
		},
		{
			name:   "different summary",
			fields: fields{Title: title, Addr: addr, Author: author, Published: puplished, Body: body},
			args: args{b: articleDB.Article{
				Title:     title,
				Addr:      addr,
				Author:    author,
				Published: puplished,
				Body:      body,
				Summary:   "A summary.",
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &articleDB.Article{
				Title:     tt.fields.Title,
				Addr:      tt.fields.Addr,
				Author:    tt.fields.Author,
				Published: tt.fields.Published,
				Body:      tt.fields.Body,
			}
			if got := a.Equal(tt.args.b); got != tt.want {
				t.Errorf("Article.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArticleService_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		log *zap.SugaredLogger
	}
	type args struct {
		ctx     context.Context
		article articleDB.Article
	}

	log, err := logger.NewTest(true)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "invalid id",
			fields: fields{log: log},
			args: args{
				ctx:     context.TODO(),
				article: articleDB.Article{ID: ""},
			},
			wantErr: true,
		},
		{
			name:   "pass",
			fields: fields{log: log},
			args: args{
				ctx:     context.TODO(),
				article: articleDB.Article{ID: articleDB.UniqueID()},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := articleDB.NewArticleService(tt.fields.log)

			if err := s.Add(tt.args.ctx, tt.args.article); (err != nil) != tt.wantErr {
				t.Errorf("ArticleService.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
