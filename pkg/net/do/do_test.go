package do_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/merlindorin/go-shared/pkg/net/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDo(t *testing.T) {
	t.Run("should make request with default options", func(t *testing.T) {
		wantMethod := http.MethodGet
		wantURL, _ := url.Parse("http://localhost")
		matchRequest := mock.MatchedBy(func(req *http.Request) bool {
			assert.Equal(t, req.Method, wantMethod)
			assert.Equal(t, req.URL, wantURL)
			return true
		})

		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(matchRequest).Return(nil, nil).Once()

		_ = do.Do(context.TODO(), wantURL, do.WithClient(mockClient))
	})

	t.Run("should return an error if the request cannot be build", func(t *testing.T) {
		assert.ErrorContains(t, do.Do(context.TODO(), &url.URL{}), "net/http: nil Context")
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
		})).Return(nil, nil)

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
		wantURL, _ := url.Parse("https://some.new")
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

	t.Run("should return an error if the response cannot be processed", func(t *testing.T) {
		wantErr := fmt.Errorf("cannot process res")
		mockClient := do.NewMockHttpClientDoer(t)
		mockClient.EXPECT().Do(mock.Anything).Return(nil, nil)

		err := do.Do(
			context.TODO(),
			&url.URL{},
			do.WithClient(mockClient),
			do.WithPostRequestHandler("mock", func(_ context.Context, _ *http.Request, _ *http.Response) error {
				return wantErr
			}),
		)

		assert.Equal(t, err, wantErr)
	})
}
