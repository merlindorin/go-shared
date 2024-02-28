package discover

import "time"

// Option represents a configuration setting that can be applied to a Discover instance.
// It allows customization of the discovery process, such as setting a timeout.
type Option func(discover *Discover)

// apply sets the given Option to the Discover instance.
func (p Option) apply(d *Discover) {
	p(d)
}

// WithTimeout sets a timeout for the discovery process.
func WithTimeout(t time.Duration) Option {
	return func(d *Discover) {
		d.timeout = t
	}
}
