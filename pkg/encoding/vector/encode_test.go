package vector

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Br0ce/articleDB/pkg/encoding"
	"github.com/Br0ce/articleDB/pkg/logger"
	"github.com/Br0ce/articleDB/pkg/vector"
)

func TestEncoder_Encode(t *testing.T) {
	type args struct {
		ctx   context.Context
		texts []string
	}
	log := logger.NewTest(false)
	tests := []struct {
		name    string
		log     *slog.Logger
		args    args
		svrFn   http.HandlerFunc
		want    []vector.Vector
		wantErr bool
	}{
		{
			name: "pass",
			log:  log,
			args: args{
				ctx:   context.TODO(),
				texts: []string{"One text", "Another text"},
			},
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := embedResponse{
					Outputs: []container[float32]{
						{
							Data: []float32{0.1, 0.2},
						},
					},
				}
				bb, err := encoding.EncodeJSON(resp)
				if err != nil {
					t.Fatalf("cannat encode test response, %s", err.Error())
				}
				w.Header().Set("Content-type", "application/json")
				_, err = w.Write(bb)
				if err != nil {
					t.Fatalf("cannot write test response")
				}
			}),
			want: []vector.Vector{
				{
					ID:   "",
					Data: []float32{0.1, 0.2},
				}},
			wantErr: false,
		},
		{
			name: "no Content-type header",
			log:  log,
			args: args{
				ctx:   context.TODO(),
				texts: []string{"One text", "Another text"},
			},
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := embedResponse{
					Outputs: []container[float32]{
						{
							Data: []float32{0.1, 0.2},
						},
					},
				}
				bb, err := encoding.EncodeJSON(resp)
				if err != nil {
					t.Fatalf("cannat encode test response, %s", err.Error())
				}
				_, err = w.Write(bb)
				if err != nil {
					t.Fatalf("cannot write test response")
				}
			}),
			want:    []vector.Vector{},
			wantErr: false,
		},
		{
			name: "server error",
			log:  log,
			args: args{
				ctx:   context.TODO(),
				texts: []string{"One text", "Another text"},
			},
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			}),
			want:    []vector.Vector{},
			wantErr: true,
		},
		{
			name: "shortcut if no texts given",
			log:  log,
			args: args{
				ctx:   context.TODO(),
				texts: []string{},
			},
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("should not be called")
			}),
			want:    []vector.Vector{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(tt.svrFn)
			enc := &Encoder{
				inferAddr: svr.URL,
				log:       tt.log,
			}

			got, err := enc.Encode(tt.args.ctx, tt.args.texts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encoder.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncoder_Ready(t *testing.T) {
	tests := []struct {
		name  string
		svrFn http.HandlerFunc
		want  bool
	}{
		{
			name: "OK",
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			}),
			want: true,
		},
		{
			name: "Not ready",
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			}),
			want: false,
		},
		{
			name: "success",
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
			}),
			want: true,
		},
		{
			name: "Bad request",
			svrFn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(401)
			}),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(tt.svrFn)
			enc := &Encoder{
				readyAddr: svr.URL,
			}
			if got := enc.Ready(); got != tt.want {
				t.Errorf("Encoder.Ready() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeInvalidBytes(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
		want  []string
	}{
		{
			name:  "empty texts",
			texts: []string{""},
			want:  []string{""},
		},
		{
			name:  "no change",
			texts: []string{"valid1", "valid2", "valid3"},
			want:  []string{"valid1", "valid2", "valid3"},
		},
		{
			name:  "change invalid",
			texts: []string{"valid1", "\xffvalid2", "valid3"},
			want:  []string{"valid1", "valid2", "valid3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeInvalidBytes(tt.texts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toValidTexts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseVectors(t *testing.T) {
	tests := []struct {
		name string
		resp embedResponse
		want []vector.Vector
	}{
		{
			name: "one embedding",
			resp: embedResponse{
				ModelName:    "model1",
				ModelVersion: "1",
				Outputs: []container[float32]{
					{
						Name:     "embedding",
						Shape:    []int{2},
						DataType: "FP32",
						Data:     []float32{0.11, 0.22},
					},
				},
			},
			want: []vector.Vector{
				{
					ID:   "",
					Data: []float32{0.11, 0.22},
				},
			},
		},
		{
			name: "multiple embeddings",
			resp: embedResponse{
				ModelName:    "model1",
				ModelVersion: "1",
				Outputs: []container[float32]{
					{
						Name:     "embedding",
						Shape:    []int{2},
						DataType: "FP32",
						Data:     []float32{0.11, 0.22},
					},
					{
						Name:     "embedding",
						Shape:    []int{2},
						DataType: "FP32",
						Data:     []float32{0.11, 0.22},
					},
				},
			},
			want: []vector.Vector{
				{
					ID:   "",
					Data: []float32{0.11, 0.22},
				},
				{
					ID:   "",
					Data: []float32{0.11, 0.22},
				},
			},
		},
		{
			name: "no embeddings",
			resp: embedResponse{
				ModelName:    "model1",
				ModelVersion: "1",
				Outputs:      []container[float32]{},
			},
			want: []vector.Vector{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseVectors(tt.resp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEmbedRequest(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
		want  embedRequest
	}{
		{
			name:  "one text",
			texts: []string{"One short sentence."},
			want: embedRequest{
				Inputs: []container[string]{
					{
						Name:     "sentences",
						Shape:    []int{1},
						DataType: "BYTES",
						Data:     []string{"One short sentence."},
					},
				},
			},
		},
		{
			name: "multiple text",
			texts: []string{
				"One short sentence.",
				"A texts",
				"Some more",
			},
			want: embedRequest{
				Inputs: []container[string]{
					{
						Name:     "sentences",
						Shape:    []int{3},
						DataType: "BYTES",
						Data: []string{
							"One short sentence.",
							"A texts",
							"Some more",
						},
					},
				},
			},
		},
		{
			name:  "no text",
			texts: []string{},
			want: embedRequest{
				Inputs: []container[string]{
					{
						Name:     "sentences",
						Shape:    []int{0},
						DataType: "BYTES",
						Data:     []string{},
					},
				},
			},
		},
		{
			name:  "empty text",
			texts: []string{""},
			want: embedRequest{
				Inputs: []container[string]{
					{
						Name:     "sentences",
						Shape:    []int{1},
						DataType: "BYTES",
						Data:     []string{""},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEmbedRequest(tt.texts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEmbedRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
