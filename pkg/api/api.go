package api

import (
	"log/slog"
	"net/http"

	"github.com/Br0ce/articleDB/pkg/adder"
	"github.com/Br0ce/articleDB/pkg/db/inmem"
	"github.com/Br0ce/articleDB/pkg/extract/noop"
)

type Api struct {
	handler http.Handler
	log     *slog.Logger
}

func New(log *slog.Logger) (*Api, error) {
	db := inmem.NewArticle()
	noop := noop.Client{}

	_, err := adder.New(
		adder.WithSummarizer(noop),
		adder.WithNamedEntityRecognizer(noop),
		adder.WithDB(db),
		adder.WithLogger(log.With("name", "api")),
	)
	if err != nil {
		return nil, err
	}

	return &Api{log: log}, nil
}

func (a *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.ServeHTTP(w, r)
}
