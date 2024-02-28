package ssdp

import (
	"context"
	"log"
	"strings"

	"github.com/koron/go-ssdp"
	"golang.org/x/sync/errgroup"
)

const (
	defaultWaitSecond = 2
	defaultRetry      = 1
)

// Transformer is a function type that takes a pointer to an ssdp.Service
// and returns an interface{} representation along with any error encountered
// during the transformation process.
type Transformer func(entry *ssdp.Service) (interface{}, error)

// Resolver represents an SSDP resolver that can
// discover services advertised over SSDP.
type Resolver struct {
	transform  Transformer // Function to transform service entries into a custom form.
	logger     *log.Logger // Logger for SSDP protocol debug messages.
	waitSecond int         // Time in seconds to wait for SSDP responses.
	retry      int         // Number of times to retry the SSDP search.
	address    string      // Local IP address to use for the SSDP search.
}

// New creates a new SSDP Resolver with optional configurations applied.
func New(opts ...Option) *Resolver {
	m := &Resolver{}

	defaultOptions := []Option{WithWaitSecond(defaultWaitSecond), WithRetry(defaultRetry)}

	opts = append(defaultOptions, opts...)

	for _, opt := range opts {
		opt.apply(m)
	}

	return m
}

// Resolve performs the SSDP discovery for the predefined settings,
// sending each discovered and optionally transformed service entry
// to the 'discovered' channel until the context is done or an error occurs.
func (m Resolver) Resolve(ctx context.Context, discovered chan<- interface{}) error {
	ch := make(chan *ssdp.Service)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(discover(ctx, m.logger, m.retry, m.waitSecond, m.address, ch))
	g.Go(process(m.transform, discovered, ch))

	return g.Wait()
}

// process receives entries from an SSDP search and applies the transform function
// to each one before sending them to the 'discovered' channel.
func process(transform Transformer, discovered chan<- interface{}, ch <-chan *ssdp.Service) func() error {
	return func() error {
		defer close(discovered)

		for entry := range ch {
			var err error
			var transformed interface{} = entry

			if transform != nil {
				transformed, err = transform(entry)
				if err != nil {
					return err
				}
			}

			discovered <- transformed
		}

		return nil
	}
}

// discover performs an SSDP search using the provided settings and sends each matching
// service entry to the given channel.
func discover(
	ctx context.Context,
	logger *log.Logger,
	retry int,
	waitSecond int,
	address string,
	ch chan<- *ssdp.Service,
) func() error {
	return func() error {
		defer close(ch)

		if logger != nil {
			ssdp.Logger = logger
		}

		for i := retry; i > 0; i-- {
			select {
			case <-ctx.Done():
				return nil
			default:
				if err := search(waitSecond, address, ch); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func search(waitSecond int, address string, ch chan<- *ssdp.Service) error {
	list, err := ssdp.Search(ssdp.RootDevice, waitSecond, address)
	if err != nil {
		return err
	}

	for _, srv := range list {
		s := srv

		if strings.Contains(srv.Server, "Sonos") && strings.Contains(srv.USN, "RINCON") {
			ch <- &s
		}
	}
	return nil
}
