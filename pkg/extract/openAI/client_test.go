package openai

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/encoding"
	"github.com/Br0ce/articleDB/pkg/logger"
)

func TestClient_Summarize(t *testing.T) {
	t.Parallel()
	type fields struct {
		apiKey string
		log    *slog.Logger
	}
	type args struct {
		ctx  context.Context
		text string
	}

	log := logger.NewTest(false)

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

func TestClient_resultText(t *testing.T) {
	t.Parallel()
	log := logger.NewTest(false)

	text := "some text"
	tests := []struct {
		name    string
		log     *slog.Logger
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

	log := logger.NewTest(false)

	tests := []struct {
		name    string
		text    string
		log     *slog.Logger
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
