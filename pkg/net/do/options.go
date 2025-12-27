package do

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"go.uber.org/zap"
)

// Params holds configuration for an HTTP request.
type Params struct {
	Client HTTPClientDoer

	Method string
	Path   string
	Body   io.Reader

	PreRequestHandlers  map[string]PreRequestHandlerFunc
	PostRequestHandlers map[string]PostRequestHandlerFunc
	ErrorHandlers       map[string]ErrorHandlerFunc

	Logger *zap.Logger

	now func() time.Time
}

// HTTPClientDoer is an interface for executing HTTP requests.
type HTTPClientDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewParams creates a new Params with initialized handler maps.
func NewParams() *Params {
	return &Params{
		PreRequestHandlers:  map[string]PreRequestHandlerFunc{},
		PostRequestHandlers: map[string]PostRequestHandlerFunc{},
		ErrorHandlers:       map[string]ErrorHandlerFunc{},
	}
}

// PreRequestHandlerFunc modifies an HTTP request before it is sent.
type PreRequestHandlerFunc func(ctx context.Context, res *http.Request) error

// Apply executes the handler on the request.
func (receiver PreRequestHandlerFunc) Apply(ctx context.Context, req *http.Request) error {
	return receiver(ctx, req)
}

// PostRequestHandlerFunc processes an HTTP response after it is received.
type PostRequestHandlerFunc func(ctx context.Context, req *http.Request, res *http.Response) error

// Apply executes the handler on the request and response.
func (receiver PostRequestHandlerFunc) Apply(ctx context.Context, req *http.Request, res *http.Response) error {
	return receiver(ctx, req, res)
}

// ErrorHandlerFunc handles errors from post-request handlers.
type ErrorHandlerFunc func(ctx context.Context, req *http.Request, res *http.Response, err error) error

// Apply executes the error handler.
func (receiver ErrorHandlerFunc) Apply(ctx context.Context, req *http.Request, res *http.Response, err error) error {
	return receiver(ctx, req, res, err)
}

// Option configures Params.
type Option func(params *Params)

// Apply applies the option to the Params.
func (p Option) Apply(params *Params) {
	p(params)
}

// WithLogger sets the logger.
func WithLogger(logger *zap.Logger) Option {
	return func(params *Params) {
		params.Logger = logger
	}
}

// WithNow sets the time function for timing and testing.
func WithNow(fn func() time.Time) Option {
	return func(params *Params) {
		params.now = fn
	}
}

// WithClient sets the HTTP client.
func WithClient(cl HTTPClientDoer) Option {
	return func(params *Params) {
		params.Client = cl
	}
}

// WithMethod sets the HTTP method (GET, POST, etc.).
func WithMethod(method string) Option {
	return func(params *Params) {
		params.Method = method
	}
}

// WithQuery adds a query parameter to the request URL.
func WithQuery(key, value string) Option {
	return WithPreRequestHandler(
		fmt.Sprintf("http_request_set_query_%s", key),
		func(_ context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Add(key, value)
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
}

// WithExtraHeader sets a single HTTP header.
func WithExtraHeader(key, value string) Option {
	return WithPreRequestHandler(
		fmt.Sprintf("http_request_set_header_%s", key),
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(key, value)
			return nil
		},
	)
}

// WithExtraHeaderf sets a single HTTP header with a formatted value.
func WithExtraHeaderf(key, format string, a ...any) Option {
	return WithExtraHeader(key, fmt.Sprintf(format, a...))
}

// WithHeader sets multiple HTTP headers from an http.Header map.
func WithHeader(header http.Header) Option {
	return WithPreRequestHandler(
		"http_request_set_header",
		func(_ context.Context, req *http.Request) error {
			for key, strings := range header {
				for _, str := range strings {
					req.Header.Set(key, str)
				}
			}
			return nil
		},
	)
}

// WithContentLength sets the Content-Length header.
func WithContentLength(requestContent []byte) Option {
	return WithPreRequestHandler(
		"http_request_content_length",
		func(_ context.Context, req *http.Request) error {
			req.ContentLength = int64(len(requestContent))
			return nil
		},
	)
}

// WithPath sets the request path with optional fmt.Sprintf formatting.
func WithPath(path string, a ...any) Option {
	return func(params *Params) {
		params.Path = fmt.Sprintf(path, a...)
	}
}

// WithBody sets the request body.
func WithBody(b io.Reader) Option {
	return func(params *Params) {
		params.Body = b
	}
}

// WithMarshalBody marshals the value to JSON and sets it as the request body.
func WithMarshalBody(v any) Option {
	return WithPreRequestHandler(
		"http_request_body_json_unmarshal",
		func(_ context.Context, req *http.Request) error {
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}

			req.Body = io.NopCloser(bytes.NewReader(b))

			return nil
		},
	)
}

// WithPreRequestHandler registers a named pre-request handler.
func WithPreRequestHandler(name string, f PreRequestHandlerFunc) Option {
	return func(params *Params) {
		params.PreRequestHandlers[name] = f
	}
}

// WithJSONRequest sets Content-Type to "application/json".
func WithJSONRequest() Option {
	return WithPreRequestHandler(
		"http_request_header_json",
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Content-Type", "application/json")
			return nil
		},
	)
}

// WithPostRequestHandler registers a named post-request handler.
func WithPostRequestHandler(name string, f PostRequestHandlerFunc) Option {
	return func(params *Params) {
		params.PostRequestHandlers[name] = f
	}
}

// WithErrorHandler registers a named error handler.
func WithErrorHandler(name string, f ErrorHandlerFunc) Option {
	return func(params *Params) {
		params.ErrorHandlers[name] = f
	}
}

// WithUnmarshalBody unmarshals the JSON response body into the provided value.
func WithUnmarshalBody(v any) Option {
	return WithPostRequestHandler(
		"http_response_body_json_unmarshal",
		func(_ context.Context, _ *http.Request, res *http.Response) error {
			if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
				return nil
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			err = json.Unmarshal(body, v)
			if err != nil {
				return err
			}

			res.Body = io.NopCloser(bytes.NewBuffer(body))
			return nil
		},
	)
}
