package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/merlindorin/go-shared/pkg/net/do"
)

// WithOtelhttp wraps the HTTP client transport with OpenTelemetry instrumentation.
// This automatically creates spans for outgoing HTTP requests with standard semantic conventions.
// Additional otelhttp options can be passed to customize the behavior.
func WithOtelhttp(opts ...otelhttp.Option) do.Option {
	return func(params *do.Params) {
		var transport = http.DefaultTransport

		if client, ok := params.Client.(*http.Client); ok && client.Transport != nil {
			transport = client.Transport
		}

		params.Client = &http.Client{
			Transport: otelhttp.NewTransport(transport, opts...),
		}
	}
}
