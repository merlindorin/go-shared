package ssdp

import "log"

// Option represents a configuration setting that can be applied to an SSDPResolver.
type Option func(resolver *Resolver)

// apply sets the given Option to the SSDPResolver.
func (o Option) apply(r *Resolver) {
	o(r)
}

// WithTransform applies a Transformer function to the SSDPResolver.
// This allows custom processing of service entries discovered via SSDP.
func WithTransform(t Transformer) Option {
	return func(resolver *Resolver) {
		resolver.transform = t
	}
}

// WithLogger sets a logger for the SSDPResolver to log SSDP protocol debug messages.
func WithLogger(l *log.Logger) Option {
	return func(resolver *Resolver) {
		resolver.logger = l
	}
}

// WithWaitSecond sets the time in seconds to wait for SSDP responses.
func WithWaitSecond(s int) Option {
	return func(resolver *Resolver) {
		resolver.waitSecond = s
	}
}

// WithRetry sets the number of times to retry the SSDP search.
func WithRetry(retry int) Option {
	return func(resolver *Resolver) {
		resolver.retry = retry
	}
}

// WithLocalAddress sets the local IP address for the SSDPResolver to use for the SSDP search.
func WithLocalAddress(l string) Option {
	return func(resolver *Resolver) {
		resolver.address = l
	}
}
