package discover

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

// Discoverer is an interface that any service discovery mechanism must implement.
// It defines a single method, Discover, to perform the discovery operation.
type Discoverer interface {
	Discover(ctx context.Context, discovered chan<- interface{}) error
}

// Resolverer is an interface that defines the Resolve method.
// Any resolver type that can implement this method is compatible with the Discover type.
type Resolverer interface {
	Resolve(ctx context.Context, discovered chan<- interface{}) error
}

// Discover represents a service discovery process which delegates the actual
// resolution work to a Resolverer implementation. It adds configurable options such as a timeout.
type Discover struct {
	resolver Resolverer    // The underlying resolver used for service discovery.
	timeout  time.Duration // A timeout for the discover process.
}

// NewDiscover creates a new instance of Discover with the provided Resolverer.
// It applies a default timeout of 1 second unless overridden by an Option provided to this function.
func NewDiscover(resolver Resolverer, opts ...Option) *Discover {
	defaultOptions := []Option{WithTimeout(time.Second)}
	opts = append(defaultOptions, opts...)

	d := &Discover{resolver: resolver}

	for _, opt := range opts {
		opt.apply(d)
	}

	return d
}

// Discover initiates the discovery process using the underlying resolver.
// It creates a context with the configured timeout and starts the resolver's Resolve method in a new goroutine.
func (d Discover) Discover(ctx context.Context, discovered chan<- interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)
	group.Go(discover(ctx, d.resolver, discovered))

	return group.Wait()
}

// discover is a helper function that calls the Resolve method of a Resolverer implementation.
// It's designed to run as a goroutine concurrently with other operations and will stop when the context expires.
func discover(ctx context.Context, resolver Resolverer, discovered chan<- interface{}) func() error {
	return func() error {
		return resolver.Resolve(ctx, discovered)
	}
}
