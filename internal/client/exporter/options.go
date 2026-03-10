package exporter

import "context"

type Option func(*Options)

type Options struct {
	Context context.Context
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
