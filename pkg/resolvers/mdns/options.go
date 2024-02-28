package mdns

import (
	"net"

	"github.com/grandcat/zeroconf"
)

// Option represents a configuration setting that can be applied to an Resolver.
type Option func(resolver *Resolver)

// apply sets the given Option to the Resolver.
func (o Option) apply(r *Resolver) {
	o(r)
}

// WithTransformer applies a Transformer function to the Resolver.
// This allows custom processing of service entries discovered via mDNS.
func WithTransformer(t Transformer) Option {
	return func(resolver *Resolver) {
		resolver.transform = t
	}
}

// WithIfaces sets the network interfaces for the Resolver to use when
// performing mDNS queries. Multiple interfaces can be provided.
func WithIfaces(ifaces ...net.Interface) Option {
	return func(resolver *Resolver) {
		resolver.ifaces = append(resolver.ifaces, ifaces...)
	}
}

// WithService sets the service name for the Resolver to discover
// over the network using the mDNS protocol.
func WithService(service string) Option {
	return func(resolver *Resolver) {
		resolver.service = service
	}
}

// WithDomain sets the domain name for the Resolver within which to perform service discovery.
func WithDomain(domain string) Option {
	return func(resolver *Resolver) {
		resolver.domain = domain
	}
}

// WithIPv4AndIPv6 instructs the Resolver to support both IPv4 and IPv6 addresses
// when resolving mDNS queries.
func WithIPv4AndIPv6() Option {
	return func(resolver *Resolver) {
		resolver.ipType = zeroconf.IPv4AndIPv6
	}
}
