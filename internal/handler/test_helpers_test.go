package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	myv "sub-service/internal/infrastructure/validator"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func newTestHandler(svc Service) *Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := myv.New()

	return &Handler{
		svc:       svc,
		logger:    logger,
		validator: validator,
	}
}

type RequestBuilder struct {
	method    string
	path      string
	body      io.Reader
	headers   map[string]string
	urlParams map[string]string
	query     url.Values
}

func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		method:    http.MethodGet,
		path:      "/",
		headers:   map[string]string{},
		urlParams: map[string]string{},
		query:     url.Values{},
	}
}

func (b *RequestBuilder) Method(m string) *RequestBuilder {
	b.method = m
	return b
}

func (b *RequestBuilder) Path(p string) *RequestBuilder {
	b.path = p
	return b
}

func (b *RequestBuilder) JSON(v interface{}) *RequestBuilder {
	data, _ := json.Marshal(v)
	b.body = bytes.NewReader(data)
	b.headers["Content-Type"] = "application/json"
	return b
}

func (b *RequestBuilder) Query(k, v string) *RequestBuilder {
	b.query.Add(k, v)
	return b
}

func (b *RequestBuilder) URLParam(k, v string) *RequestBuilder {
	b.urlParams[k] = v
	return b
}

func (b *RequestBuilder) Build() *http.Request {
	fullURL := b.path
	if len(b.query) > 0 {
		fullURL += "?" + b.query.Encode()
	}

	req := httptest.NewRequest(b.method, fullURL, b.body)

	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	if len(b.urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for k, v := range b.urlParams {
			rctx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return req
}

type Response struct {
	T   *testing.T
	Rec *httptest.ResponseRecorder
}

func NewResponse(t *testing.T) *Response {
	return &Response{
		T:   t,
		Rec: httptest.NewRecorder(),
	}
}

func (r *Response) Status(code int) *Response {
	require.Equal(r.T, code, r.Rec.Code)
	return r
}

func (r *Response) JSON(v interface{}) *Response {
	err := json.Unmarshal(r.Rec.Body.Bytes(), v)
	require.NoError(r.T, err)
	return r
}

func (r *Response) BodyContains(s string) *Response {
	require.Contains(r.T, r.Rec.Body.String(), s)
	return r
}

func Call(
	h *Handler,
	handlerFunc http.HandlerFunc,
	req *http.Request,
	resp *Response,
) {
	handlerFunc(resp.Rec, req)
}
