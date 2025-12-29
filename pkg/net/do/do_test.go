package do_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/merlindorin/go-shared/pkg/must"
	"github.com/merlindorin/go-shared/pkg/net/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDo(t *testing.T) {
	t.Run("should make request with default options", func(t *testing.T) {
		wantMethod := http.MethodGet
		wantURL := must.Get(url.Parse("http://localhost"))
		matchRequest := mock.MatchedBy(func(req *http.Request) bool {
			assert.Equal(t, req.Method, wantMethod)
			assert.Equal(t, req.URL, wantURL)
			return true
		})

		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(matchRequest).Return(&http.Response{}, nil).Once()

		_ = do.Do(context.TODO(), wantURL, do.WithClient(mockClient))
	})

	t.Run("should return an error if the request cannot be build", func(t *testing.T) {
		assert.ErrorContains(t, do.Do(context.TODO(), &url.URL{}), "unsupported protocol scheme")
	})

	t.Run("should return an error if the request cannot be made", func(t *testing.T) {
		wantErr := fmt.Errorf("cannot maje the request")

		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.Anything).Return(nil, wantErr)

		assert.ErrorIs(t, do.Do(context.TODO(), &url.URL{}, do.WithClient(mockClient)), wantErr)
	})

	t.Run("should use the same context during the execution (reqHandler, req, resHandler)", func(t *testing.T) {
		wantCtx := context.TODO()

		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			assert.Equal(t, req.Context(), wantCtx)
			return true
		})).Return(&http.Response{}, nil)

		_ = do.Do(
			wantCtx,
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPreRequestHandler("mock", func(ctx context.Context, _ *http.Request) error {
				assert.Equal(t, ctx, wantCtx)
				return nil
			}),
			do.WithPostRequestHandler("mock", func(ctx context.Context, _ *http.Request, _ *http.Response) error {
				assert.Equal(t, ctx, wantCtx)
				return nil
			}),
		)
	})

	t.Run("should be able to prepare request ", func(t *testing.T) {
		wantMethod := http.MethodOptions
		wantURL := must.Get(url.Parse("https://some.new"))
		matchRequest := mock.MatchedBy(func(req *http.Request) bool {
			assert.Equal(t, req.Method, wantMethod)
			assert.Equal(t, req.URL, wantURL)
			return true
		})

		res := &http.Response{}
		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(matchRequest).Return(res, nil).Once()

		_ = do.Do(
			context.TODO(),
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPreRequestHandler("mock", func(_ context.Context, request *http.Request) error {
				request.Method = wantMethod
				request.URL = wantURL
				return nil
			}))
	})

	t.Run("should return an error if the request cannot be prepared", func(t *testing.T) {
		wantErr := fmt.Errorf("cannot be prepared")
		preRequestHandlerMock := do.NewMockPreRequestHandlerFunc(t)
		preRequestHandlerMock.EXPECT().Execute(mock.Anything, mock.Anything).Return(wantErr).Once()

		err := do.Do(
			context.TODO(),
			&url.URL{},
			do.WithPreRequestHandler("mock", preRequestHandlerMock.Execute),
		)

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("should be able to process response ", func(t *testing.T) {
		wantRes := &http.Response{}
		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.Anything).Return(wantRes, nil)

		_ = do.Do(
			context.TODO(),
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPostRequestHandler("mock", func(_ context.Context, _ *http.Request, res *http.Response) error {
				assert.Equal(t, res, wantRes)
				return nil
			}),
		)
	})

	t.Run("should return an error if the response cannot be processed with no error handler", func(t *testing.T) {
		wantErr := fmt.Errorf("cannot process res")
		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.Anything).Return(&http.Response{Body: io.NopCloser(strings.NewReader(""))}, nil)

		err := do.Do(
			context.TODO(),
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPostRequestHandler("mock", func(_ context.Context, _ *http.Request, _ *http.Response) error {
				return wantErr
			}),
		)

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("should return no error if an error handler process it", func(t *testing.T) {
		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.Anything).Return(&http.Response{Body: io.NopCloser(strings.NewReader(""))}, nil)

		err := do.Do(
			context.TODO(),
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPostRequestHandler("mock", func(_ context.Context, _ *http.Request, _ *http.Response) error {
				return fmt.Errorf("cannot process res")
			}),
			do.WithErrorHandler("mock", func(_ context.Context, _ *http.Request, _ *http.Response, _ error) error {
				return nil
			}),
		)

		assert.NoError(t, err)
	})
}
