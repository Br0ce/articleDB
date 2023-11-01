package elastic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Engine struct {
	client *elasticsearch.Client
	log    *slog.Logger
}

type Config struct {
	Addrs    []string
	Username string
	Password string
	Cert     []byte
	Mapping  string
}

func NewEngine(ctx context.Context, cfg Config, log *slog.Logger) (*Engine, error) {
	log.Info("create new elastic engine", "method", "NewEngine")

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addrs,
		Username:  cfg.Username,
		Password:  cfg.Password,
		CACert:    cfg.Cert,
	})
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("could not connect to elastic cluster, %s", res.Status())
	}
	log.Debug("check elastic info", "method", "NewEngine", "info", res)

	e := &Engine{client: es, log: log}
	err = e.createIndex(ctx, "article-german-index-001", cfg.Mapping)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Engine) createIndex(ctx context.Context, name string, mapping string) error {
	e.log.Debug("create article index", "method", "createIndex")

	exists, err := e.indexExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		e.log.Debug("index already exits", "method", "createIndex", "indexname", name)
		return nil
	}

	resp, err := esapi.IndicesCreateRequest{
		Index: name,
		Body:  strings.NewReader(mapping),
	}.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("cannot create index=%s, %s", name, resp.Status())
	}
	e.log.Debug("created article index", "method", "createIndex", "status", resp.Status())

	return nil
}

func (e *Engine) indexExists(ctx context.Context, index string) (bool, error) {
	e.log.Debug("check if index exists", "method", "indexExists", "indexname", index)

	resp, err := esapi.IndicesExistsRequest{
		Index: []string{index},
	}.Do(ctx, e.client)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return false, nil
	}

	return true, nil
}
