package inmem

import (
	"context"
	"reflect"
	"testing"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/ids"
	"golang.org/x/sync/errgroup"
)

func TestArticle_Add(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		item article.Article
	}

	tests := []struct {
		items   map[string]article.Article
		name    string
		args    args
		want    article.Article
		wantErr bool
	}{
		{
			name:  "pass",
			items: make(map[string]article.Article),
			args: args{
				ctx: context.TODO(),
				item: article.Article{
					ID:    "",
					Title: "test title",
				},
			},

			want: article.Article{
				Title: "test title",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Article{
				items: tt.items,
			}

			id, err := a.Add(tt.args.ctx, tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("Article.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, ok := tt.items[id]
			if !ok {
				t.Error("Article.Add() = item not present")
			}

			if got.ID != id {
				t.Error("Article.Add() = item id not sett")
			}

		})
	}
}

func TestArticle_AddAndGet_parallel(t *testing.T) {
	t.Parallel()

	db := Article{items: make(map[string]article.Article)}

	eg := new(errgroup.Group)
	ctx := context.TODO()

	num := 100
	ids := make(chan string, num*2)
	defer close(ids)

	// Add num articles to db.
	for i := 0; i < num; i++ {
		id, err := db.Add(ctx, article.Article{})
		if err != nil {
			t.Fatalf("could not prime with articles")
		}
		ids <- id
	}

	for i := 0; i < num; i++ {

		// Get the num articles out.
		eg.Go(func() error {
			id := <-ids
			_, err := db.Get(ctx, id)
			if err != nil {
				t.Fatalf("could not get article, id=%s", id)
			}
			return nil
		})

		// Add num more articles.
		eg.Go(func() error {
			a := article.Article{}
			id, err := db.Add(ctx, a)
			if err != nil {
				return err
			}
			ids <- id
			return nil
		})

	}

	err := eg.Wait()
	if err != nil {
		t.Fatalf("finished with err, %s", err.Error())
	}

	// It should only be num articles present.
	if len(ids) != num {
		t.Fatalf("ids len, want %v got %v", num, len(ids))
	}

	for i := 0; i < num; i++ {
		id := <-ids
		_, ok := db.items[id]
		if !ok {
			t.Fatalf("id not present, id=%s", id)
		}
	}
}

func TestArticle_Get(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  string
	}

	id := ids.UniqueID()
	ar := article.Article{ID: id}

	items := make(map[string]article.Article)
	items[id] = ar

	tests := []struct {
		items   map[string]article.Article
		name    string
		args    args
		want    article.Article
		wantErr bool
	}{
		{
			name:  "pass",
			items: items,
			args: args{
				ctx: context.TODO(),
				id:  id,
			},
			want:    ar,
			wantErr: false,
		},
		{
			name:  "invalid id",
			items: items,
			args: args{
				ctx: context.TODO(),
				id:  "",
			},
			wantErr: true,
		},
		{
			name:  "item not found",
			items: items,
			args: args{
				ctx: context.TODO(),
				id:  ids.UniqueID(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Article{
				items: tt.items,
			}

			got, err := a.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Article.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Article.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewArticle(t *testing.T) {
	t.Parallel()

	got := NewArticle()
	if got.items == nil {
		t.Error("Article items is nil")
	}
}
