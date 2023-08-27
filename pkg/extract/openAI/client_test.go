package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/encoding"
	"github.com/Br0ce/articleDB/pkg/logger"
)

func TestClient_Summarize(t *testing.T) {
	t.Parallel()
	type fields struct {
		apiKey string
		log    *zap.SugaredLogger
	}
	type args struct {
		ctx  context.Context
		text string
	}

	log, err := logger.NewTest(false)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	text := "Some text"
	response := "response"

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "pass",
			fields: fields{
				apiKey: "some key",
				log:    log,
			},
			args: args{
				ctx:  context.TODO(),
				text: text,
			},
			want:    response,
			wantErr: false,
		},
		{
			name: "empty text",
			fields: fields{
				apiKey: "some key",
				log:    log,
			},
			args: args{
				ctx:  context.TODO(),
				text: "",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				var dto completionDTO
				err := encoding.DecodeJSON(r.Body, &dto)
				if err != nil {
					t.Fatalf("could not decode body, %s", err.Error())
				}

				if dto.Model != gpt3TextModel {
					t.Fatalf("model: want %s got %s", gpt3TextModel, dto.Model)
				}
				if dto.Prompt != fmt.Sprintf("%s\n\nTl;dr", text) {
					t.Fatalf("text: want %s got %s", text, dto.Prompt)
				}

				resp := responseDTO{
					Choices: []choicesDTO{{Text: response}},
				}
				bb, err := encoding.EncodeJSON(resp)
				if err != nil {
					t.Fatalf("could not encode, %s", err.Error())
				}

				_, err = w.Write(bb)
				if err != nil {
					t.Fatalf("could not write bytes, %s", err.Error())
				}
			}))
			defer svr.Close()

			c := &Client{
				apiKey:         tt.fields.apiKey,
				completionAddr: svr.URL,
				log:            tt.fields.log,
			}

			got, err := c.Summarize(tt.args.ctx, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Summarize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Client.Summarize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_httpRequest_pass(t *testing.T) {
	t.Parallel()
	apiKey := "testKey"
	id := "1234"
	payload := "payload"

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if a != fmt.Sprintf("Bearer %s", apiKey) {
			t.Fatalf("apiKey: want %s got %s", apiKey, a)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Fatalf("content-type: want application/json got %s", ct)
		}

		pp, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read body, %s", err.Error())
		}
		defer r.Body.Close()
		if string(pp) != payload {
			t.Fatalf("body: want %s got %s", payload, string(pp))
		}

		resp := responseDTO{ID: id}
		bb, err := encoding.EncodeJSON(resp)
		if err != nil {
			t.Fatalf("could not encode, %s", err.Error())
		}

		_, err = w.Write(bb)
		if err != nil {
			t.Fatalf("could not write bytes, %s", err.Error())
		}

	}))
	defer svr.Close()

	t.Run("pass", func(t *testing.T) {
		log, err := logger.NewTest(false)
		if err != nil {
			t.Fatalf("could not init logger, %s", err.Error())
		}

		c := &Client{
			apiKey:         apiKey,
			completionAddr: svr.URL,
			log:            log,
		}

		got, err := c.httpRequest(context.TODO(), strings.NewReader(payload))
		if err != nil {
			t.Errorf("Client.execRequest() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, responseDTO{ID: id}) {
			t.Errorf("Client.execRequest() = %v", got)
		}
	})
}

func TestClient_httpRequest_fail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		timeout     time.Duration
		handlerFunc http.HandlerFunc
	}{
		{
			name:    "response code > 299",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(300)
			}),
		},
		{
			name:    "timeout",
			timeout: time.Millisecond * 2,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 10)
			}),
		},
		{
			name:    "invalid dto",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte{})
				if err != nil {
					t.Fatalf("could not write bytes, %s", err.Error())
				}
			}),
		},
	}

	log, err := logger.NewTest(true)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(tt.handlerFunc)
			defer svr.Close()

			c := &Client{
				completionAddr: svr.URL,
				log:            log,
			}

			ctx, cancleFn := context.WithTimeout(context.TODO(), tt.timeout)
			defer cancleFn()

			_, err := c.httpRequest(ctx, strings.NewReader(""))

			if err == nil {
				t.Fatalf("Client.execRequest() without err")
			}
		})
	}
}

func TestClient_resultText(t *testing.T) {
	t.Parallel()
	log, err := logger.NewTest(false)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	text := "some text"
	tests := []struct {
		name    string
		log     *zap.SugaredLogger
		arg     responseDTO
		want    string
		wantErr bool
	}{
		{
			name: "pass",
			log:  log,
			arg: responseDTO{
				Choices: []choicesDTO{{Text: text}},
			},
			want:    text,
			wantErr: false,
		},
		{
			name:    "no result content",
			log:     log,
			arg:     responseDTO{},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid result content",
			log:  log,
			arg: responseDTO{
				Choices: []choicesDTO{{Text: ""}},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "multiple responses",
			log:  log,
			arg: responseDTO{
				Choices: []choicesDTO{
					{Text: text},
					{Text: "some other text"},
				},
			},
			want:    text,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				log: tt.log,
			}

			got, err := c.resultText(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.resultText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Client.resultText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_toNER(t *testing.T) {
	t.Parallel()

	log, err := logger.NewTest(true)
	if err != nil {
		t.Fatalf("could not init logger, %s", err.Error())
	}

	tests := []struct {
		name    string
		text    string
		log     *zap.SugaredLogger
		want    article.NER
		wantErr bool
	}{
		{
			name:    "empty text",
			text:    "",
			log:     log,
			wantErr: true,
			want:    article.NER{},
		},
		{
			name:    "pass with newlines",
			text:    "\n\n{\"Person\": [\"Gérald Darmanin\", \"Élisabeth Borne\"], \n\"Location\": [\"Frankreich\"], \n\"Organization\": [\"Polizei\"]}",
			log:     log,
			wantErr: false,
			want: article.NER{
				Pers: []string{"Gérald Darmanin", "Élisabeth Borne"},
				Locs: []string{"Frankreich"},
				Orgs: []string{"Polizei"},
			},
		},
		{
			name:    "pass",
			text:    "{\"Person\": [\"Gérald Darmanin\", \"Élisabeth Borne\"], \"Location\": [\"Frankreich\"], \"Organization\": [\"Polizei\"]}",
			log:     log,
			wantErr: false,
			want: article.NER{
				Pers: []string{"Gérald Darmanin", "Élisabeth Borne"},
				Locs: []string{"Frankreich"},
				Orgs: []string{"Polizei"},
			},
		},
		{
			name:    "pass without key",
			text:    "{ \"Location\": [\"Frankreich\"], \"Organization\": [\"Polizei\"]}",
			log:     log,
			wantErr: false,
			want: article.NER{
				Locs: []string{"Frankreich"},
				Orgs: []string{"Polizei"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				log: tt.log,
			}

			got, err := c.toNER(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.toNER() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.toNER() = %q, want %q", got, tt.want)
			}
		})
	}
}
