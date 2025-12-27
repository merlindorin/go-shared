package do

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// Do executes an HTTP request to the given URL with the provided options.
func Do(ctx context.Context, u *url.URL, options ...Option) error {
	defaultOptions := []Option{
		WithMethod(http.MethodGet),
		WithClient(http.DefaultClient),
		WithLogger(zap.NewNop()),
		WithNow(time.Now),
	}

	p := NewParams()

	for _, option := range append(defaultOptions, options...) {
		option(p)
	}

	start := p.now()
	log := p.Logger.With(zap.Time("start", start))

	log.Debug("buildRequest", zap.Duration("duration", time.Since(start)))
	req, err := http.NewRequestWithContext(ctx, p.Method, u.JoinPath(p.Path).String(), p.Body)
	if err != nil {
		log.Error("cannot buildRequest", zap.Error(err))
		return err
	}

	// Run pre-request handlers.
	for name, preRequestHandler := range p.PreRequestHandlers {
		log.Debug("preRequest", zap.Duration("duration", time.Since(start)), zap.String("preRequestHandlerName", name))
		if err = preRequestHandler.Apply(ctx, req); err != nil {
			log.Error("cannot handle request", zap.Error(err), zap.String("preRequestHandlerName", name))
			return err
		}
	}

	log.Debug("sendRequest", zap.Duration("duration", time.Since(start)))
	res, err := p.Client.Do(req)
	if err != nil {
		log.Error("cannot sendRequest", zap.Error(err))
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("cannot close Body", zap.Error(err))
		}
	}(res.Body)

	// Run post-request handlers.
	for name, postRequestHandler := range p.PostRequestHandlers {
		log.Debug("postRequest", zap.Duration("duration", time.Since(start)), zap.String("postRequestHandlerName", name))

		if err = postRequestHandler.Apply(ctx, req, res); err != nil {
			log.Info("cannot handle response", zap.Error(err), zap.String("postRequestHandlerName", name))
			continue
		}
	}

	// Run error handlers.
	for name, errorHandler := range p.ErrorHandlers {
		log.Debug("errorHandler", zap.Duration("duration", time.Since(start)), zap.String("errorHandlerName", name))

		if err = errorHandler.Apply(ctx, req, res, err); err != nil {
			log.Error("cannot handle response", zap.Error(err), zap.String("postRequestHandlerName", name))
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("cannot process request: %w", err)
	}

	return nil
}
