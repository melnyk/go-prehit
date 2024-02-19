package prehit

import "go.melnyk.org/mlog"

// Options
type options struct {
	logger  mlog.Logger
	maxsize uint
	metrics Metrics
}

// Options is a set of options for the prehit package.
type Option interface {
	apply(*options)
}

type loggerOption struct {
	logger mlog.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.logger
}

// WithLogger sets the logger for the prehit package.
func WithLogger(logger mlog.Logger) Option {
	return loggerOption{logger: logger}
}

type maxsizeOption uint

func (o maxsizeOption) apply(opts *options) {
	opts.maxsize = uint(o)
}

// WithMaxSize sets the maximum size of the cache.
func WithMaxSize(size uint) Option {
	return maxsizeOption(size)
}

type metricsOption struct {
	metrics Metrics
}

func (o metricsOption) apply(opts *options) {
	opts.metrics = o.metrics
}

// WithMetrics sets the metrics for the prehit package.
func WithMetrics(metrics Metrics) Option {
	return metricsOption{metrics: metrics}
}
