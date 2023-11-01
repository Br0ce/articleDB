package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Br0ce/articleDB/pkg/encoding"
)

var (
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrBadGateway          = errors.New("bad gateway")
	ErrInternalServer      = errors.New("internal server error")
	ErrServer              = errors.New("server error")
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
)

// Post performs a post request to addr with the content of r as body. If the response
// content-type is application/json the response body is stored in the value pointed
// to by v.
// To timeout the httpRequest, use an appropriate context.
//
// There is no retrying or throttling performed.
func Post(ctx context.Context, addr string, header http.Header, r io.Reader, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, r)
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrBadGateway)
	}
	req.Header = header
	return do(ctx, req, v)
}

// Get performs a get request to addr with the content of r as body. If the response
// content-type is application/json the response body is stored in the value pointed
// to by v.
// To timeout the httpRequest, use an appropriate context.
//
// There is no retrying or throttling performed.
func Get(ctx context.Context, addr string, header http.Header, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrBadGateway)
	}
	req.Header = header
	return do(ctx, req, v)
}

// do executes the given request. If the response content-type is
// application/json the response body is stored in the value pointed to by v.
// To timeout the httpRequest, use an appropriate context.
//
// There is no retrying or throttling performed.
func do(ctx context.Context, req *http.Request, v any) error {
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrBadGateway)
	}
	defer resp.Body.Close()

	if !statusOK(resp.StatusCode) {
		return getErr(resp.StatusCode)
	}

	if resp.Header.Get("Content-type") != "application/json" {
		return nil
	}

	err = encoding.DecodeJSON(resp.Body, &v)
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrUnprocessableEntity)
	}

	return nil
}

// statusOK checks if code is between 200 and 300.
func statusOK(code int) bool {
	return code >= 200 && code < 300
}

// getErr returns ErrBadRequest or ErrInternalServer depending on the given code.
func getErr(code int) error {
	if code == 401 {
		return ErrUnauthorized
	}
	if code == 403 {
		return ErrForbidden
	}
	if code == 404 {
		return ErrNotFound
	}
	if code == 422 {
		return ErrUnprocessableEntity
	}
	if code >= 400 && code < 500 {
		return ErrBadRequest
	}
	if code == 500 {
		return ErrInternalServer
	}
	if code == 502 {
		return ErrBadGateway
	}
	if code >= 500 {
		return ErrServer
	}
	return fmt.Errorf("code: %d", code)
}
