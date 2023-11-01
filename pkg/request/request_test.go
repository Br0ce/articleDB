package request

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Br0ce/articleDB/pkg/encoding"
)

func TestPost_pass(t *testing.T) {
	t.Parallel()

	type dto struct {
		ID string `json:"id"`
	}

	id := "1234"
	payload := "payload"
	header := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer 1324"},
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expect post method, got %s", r.Method)
		}

		for k, v := range header {
			if v[0] != r.Header.Get(k) {
				t.Fatalf("header %s is not equal, want %s got %s", k, v, r.Header.Get(k))
			}
		}

		pp, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read body, %s", err.Error())
		}
		defer r.Body.Close()
		if string(pp) != payload {
			t.Fatalf("body: want %s got %s", payload, string(pp))
		}

		bb, err := encoding.EncodeJSON(dto{ID: id})
		if err != nil {
			t.Fatalf("could not encode, %s", err.Error())
		}

		w.Header().Add("Content-type", "application/json")
		_, err = w.Write(bb)
		if err != nil {
			t.Fatalf("could not write bytes, %s", err.Error())
		}
	}))
	defer svr.Close()

	t.Run("pass", func(t *testing.T) {
		var got dto
		err := Post(context.TODO(), svr.URL, header, strings.NewReader(payload), &got)
		if err != nil {
			t.Errorf("not expected error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, dto{ID: id}) {
			t.Errorf("not Equal, got = %v", got)
		}
	})
}

func TestGet_pass(t *testing.T) {
	t.Parallel()

	type dto struct {
		ID string `json:"id"`
	}

	id := "1234"
	header := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer 1324"},
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expect get method, got %s", r.Method)
		}

		for k, v := range header {
			if v[0] != r.Header.Get(k) {
				t.Fatalf("header %s is not equal, want %s got %s", k, v, r.Header.Get(k))
			}
		}

		bb, err := encoding.EncodeJSON(dto{ID: id})
		if err != nil {
			t.Fatalf("could not encode, %s", err.Error())
		}

		w.Header().Add("Content-type", "application/json")
		_, err = w.Write(bb)
		if err != nil {
			t.Fatalf("could not write bytes, %s", err.Error())
		}
	}))
	defer svr.Close()

	t.Run("pass", func(t *testing.T) {
		var got dto
		err := Get(context.TODO(), svr.URL, header, &got)
		if err != nil {
			t.Errorf("not expected error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, dto{ID: id}) {
			t.Errorf("not Equal, got = %v", got)
		}
	})
}

func TestDo_Fail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		timeout     time.Duration
		handlerFunc http.HandlerFunc
		header      http.Header
		value       any
		wantErr     error
	}{
		{
			name:    "Unauthorized",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(401)
			}),
			value:   struct{}{},
			wantErr: ErrUnauthorized,
		},
		{
			name:    "Forbidden",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(403)
			}),
			value:   struct{}{},
			wantErr: ErrForbidden,
		},
		{
			name:    "BadRequest",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			}),
			value:   struct{}{},
			wantErr: ErrBadRequest,
		},
		{
			name:    "InternalServerErr",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			}),
			value:   struct{}{},
			wantErr: ErrInternalServer,
		},
		{
			name:    "timeout",
			timeout: time.Millisecond * 2,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 10)
			}),
			value:   struct{}{},
			wantErr: ErrBadGateway,
		},
		{
			name:    "invalid dto",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-type", "application/json")
				_, err := w.Write([]byte{})
				if err != nil {
					t.Fatalf("could not write bytes, %s", err.Error())
				}
			}),
			value: struct {
				ID string `json:"id"`
			}{},
			wantErr: ErrUnprocessableEntity,
		},
		{
			name:    "response header with application/json, but no value provided ",
			timeout: time.Second,
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-type", "application/json")
				_, err := w.Write([]byte{})
				if err != nil {
					t.Fatalf("could not write bytes, %s", err.Error())
				}
			}),
			value:   nil,
			wantErr: ErrUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(tt.handlerFunc)
			defer svr.Close()
			ctx, cancleFn := context.WithTimeout(context.TODO(), tt.timeout)
			defer cancleFn()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
			if err != nil {
				t.Fatalf("cannot create test request, %s", err.Error())
			}
			req.Header = tt.header

			err = do(context.TODO(), req, tt.value)
			if err == nil {
				t.Fatalf("err is nil")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: want %s got %s", tt.wantErr, err.Error())
			}
		})
	}
}

func Test_getErr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		code    int
		wantErr error
	}{
		{
			name:    "BadRequest",
			code:    400,
			wantErr: ErrBadRequest,
		},
		{
			name:    "Unauthorized",
			code:    401,
			wantErr: ErrUnauthorized,
		},
		{
			name:    "Forbidden",
			code:    403,
			wantErr: ErrForbidden,
		},
		{
			name:    "NotFound",
			code:    404,
			wantErr: ErrNotFound,
		},
		{
			name:    "PreconditionFailed",
			code:    412,
			wantErr: ErrBadRequest,
		},
		{
			name:    "UnprocessableEntity",
			code:    422,
			wantErr: ErrUnprocessableEntity,
		},
		{
			name:    "InternalServerErr",
			code:    500,
			wantErr: ErrInternalServer,
		},
		{
			name:    "BadGateway",
			code:    502,
			wantErr: ErrBadGateway,
		},
		{
			name:    "ServiceUnavailable",
			code:    503,
			wantErr: ErrServer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getErr(tt.code)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("getErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
