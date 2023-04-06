package test

import (
	"net/url"
	"testing"
	"time"

	articleDB "github.com/Br0ce/articleDB/pkg"
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
