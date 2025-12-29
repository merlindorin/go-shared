package rest_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/merlindorin/go-shared/pkg/must"
	"github.com/merlindorin/go-shared/pkg/net/do"
	"github.com/merlindorin/go-shared/pkg/net/rest"
)

func TestNewRest(t *testing.T) {
	t.Run("should have do", func(t *testing.T) {
		wantURL := must.Get(url.Parse("https://merlindorin.com"))
		wantMethod := http.MethodGet
		matchRequest := mock.MatchedBy(func(req *http.Request) bool {
			assert.Equal(t, req.Method, wantMethod)
			assert.Equal(t, req.URL, wantURL)
			return true
		})

		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(matchRequest).Return(&http.Response{Body: io.NopCloser(strings.NewReader(""))}, nil).Once()

		r := rest.NewRest(wantURL, do.WithClient(mockClient))
		err := r.Do(t.Context())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("should have predefined methods", func(t *testing.T) {
		wantURL := must.Get(url.Parse("https://merlindorin.com"))
		tests := []struct {
			name       string
			method     string
			callMethod func(r *rest.Rest) error
		}{
			{"GET", http.MethodGet, func(r *rest.Rest) error { return r.GET(t.Context()) }},
			{"POST", http.MethodPost, func(r *rest.Rest) error { return r.POST(t.Context()) }},
			{"PUT", http.MethodPut, func(r *rest.Rest) error { return r.PUT(t.Context()) }},
			{"PATCH", http.MethodPatch, func(r *rest.Rest) error { return r.PATCH(t.Context()) }},
			{"DELETE", http.MethodDelete, func(r *rest.Rest) error { return r.DELETE(t.Context()) }},
			{"OPTIONS", http.MethodOptions, func(r *rest.Rest) error { return r.OPTIONS(t.Context()) }},
			{"HEAD", http.MethodHead, func(r *rest.Rest) error { return r.HEAD(t.Context()) }},
			{"CONNECT", http.MethodConnect, func(r *rest.Rest) error { return r.CONNECT(t.Context()) }},
			{"TRACE", http.MethodTrace, func(r *rest.Rest) error { return r.TRACE(t.Context()) }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				matchRequest := mock.MatchedBy(func(req *http.Request) bool {
					assert.Equal(t, tt.method, req.Method)
					return true
				})

				mockClient := do.NewMockHttpClientDoer(t)
				mockClient.EXPECT().Do(matchRequest).Return(&http.Response{}, nil).Once()

				r := rest.NewRest(wantURL, do.WithClient(mockClient))
				err := tt.callMethod(r)

				if err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}
