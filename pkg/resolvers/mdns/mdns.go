package mdns

import (
	"context"
	"fmt"
	"net"

	"github.com/grandcat/zeroconf"
	"golang.org/x/sync/errgroup"
)

// Transformer is a function type that takes a pointer to a zeroconf.ServiceEntry
// and returns an interface{} representation along with any error encountered during
// the transformation process.
type Transformer func(entry *zeroconf.ServiceEntry) (interface{}, error)

// Resolver represents a multicast DNS resolver that can
// discover services advertised over mDNS.
type Resolver struct {
	ifaces  []net.Interface // Local network interfaces to use for the mDNS queries.
	ipType  zeroconf.IPType // The IP protocol version to use (IPv4, IPv6, or both).
	service string          // The service name to look for.
	domain  string          // The domain in which to look for the service.

	transform Transformer // Function to transform service entries into a custom form.
}

// New creates a new MDNS Resolver with optional configurations applied.
func New(opts ...Option) *Resolver {
	m := &Resolver{
		ipType: zeroconf.IPv4AndIPv6,
	}

	for _, opt := range opts {
		opt.apply(m)
	}

	return m
}

// Resolve performs the mDNS query for the resolver's configured service and domain,
// sending each discovered and optionally transformed service entry to the 'discovered'
// channel until the context is done or an error occurs.
func (m Resolver) Resolve(ctx context.Context, discovered chan<- interface{}) error {
	defer close(discovered)

	resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces(m.ifaces), zeroconf.SelectIPTraffic(m.ipType))
	if err != nil {
		return fmt.Errorf("cannot create new mdns resolver: %w", err)
	}

	chEntries := make(chan *zeroconf.ServiceEntry)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return resolver.Browse(ctx, m.service, m.domain, chEntries)
	})

	g.Go(func() error {
		for entry := range chEntries {
			var transformed interface{} = entry

			if m.transform != nil {
				transformed, err = m.transform(entry)
				if err != nil {
					return err
				}
			}
			discovered <- transformed
		}

		return nil
	})

	return g.Wait()
}
