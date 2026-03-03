package repo

import "context"

type Option func(*Options)

type Options struct {
	Location string
	Context  context.Context
}

func WithLocation(location string) Option {
	return func(o *Options) {
		o.Location = location
	}
}

func NewOptions(opts ...Option) Options {
	opt := Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}
