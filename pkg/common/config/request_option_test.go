package config

import (
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

// TestRequestOptions test request options with custom values
func TestRequestOptions(t *testing.T) {
	opt := NewRequestOptions([]RequestOption{
		WithTag("a", "b"),
		WithTag("c", "d"),
		WithTag("e", "f"),
		WithSD(true),
		WithDialTimeout(time.Second),
		WithReadTimeout(time.Second),
		WithWriteTimeout(time.Second),
	})
	assert.DeepEqual(t, "b", opt.Tag("a"))
	assert.DeepEqual(t, "d", opt.Tag("c"))
	assert.DeepEqual(t, "f", opt.Tag("e"))
	assert.DeepEqual(t, time.Second, opt.DialTimeout())
	assert.DeepEqual(t, time.Second, opt.ReadTimeout())
	assert.DeepEqual(t, time.Second, opt.WriteTimeout())
	assert.True(t, opt.IsSD())
}

// TestRequestOptionsWithDefaultOpts test request options with default values
func TestRequestOptionsWithDefaultOpts(t *testing.T) {
	SetPreDefinedOpts(WithTag("pre-defined", "blablabla"), WithTag("a", "default-value"), WithSD(true))
	opt := NewRequestOptions([]RequestOption{
		WithTag("a", "b"),
		WithSD(false),
	})
	assert.DeepEqual(t, "b", opt.Tag("a"))
	assert.DeepEqual(t, "blablabla", opt.Tag("pre-defined"))
	assert.DeepEqual(t, map[string]string{
		"a":           "b",
		"pre-defined": "blablabla",
	}, opt.Tags())
	assert.False(t, opt.IsSD())
	SetPreDefinedOpts()
	assert.Nil(t, preDefinedOpts)
	assert.DeepEqual(t, time.Duration(0), opt.WriteTimeout())
	assert.DeepEqual(t, time.Duration(0), opt.ReadTimeout())
	assert.DeepEqual(t, time.Duration(0), opt.DialTimeout())
}

// TestRequestOptions_CopyTo test request options copy to another one
func TestRequestOptions_CopyTo(t *testing.T) {
	opt := NewRequestOptions([]RequestOption{
		WithTag("a", "b"),
		WithSD(false),
	})
	var copyOpt RequestOptions
	opt.CopyTo(&copyOpt)
	assert.DeepEqual(t, opt.Tags(), copyOpt.Tags())
	assert.DeepEqual(t, opt.IsSD(), copyOpt.IsSD())
}
