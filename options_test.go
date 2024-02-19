package prehit

import (
	"testing"

	"go.melnyk.org/mlog/nolog"
)

func TestWithLogger(t *testing.T) {
	logger := nolog.NewLogbook().Joiner().Join("")
	o := WithLogger(logger)

	local := &options{}
	o.apply(local)

	if local.logger != logger {
		t.Error("Expected logger to be set")
	}
}

// Unittest for WithMaxSize
func TestWithMaxSize(t *testing.T) {
	o := WithMaxSize(10)

	local := &options{}
	o.apply(local)

	if local.maxsize != 10 {
		t.Error("Expected maxSize to be set")
	}
}

func TestWithMetrics(t *testing.T) {
	metrics := &nometrics{}
	o := WithMetrics(metrics)

	local := &options{}
	o.apply(local)

	if local.metrics != metrics {
		t.Error("Expected metrics to be set")
	}
}
