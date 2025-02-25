package config

import (
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

// TestDefaultClientOptions test client options with default values
func TestDefaultClientOptions(t *testing.T) {
	options := NewClientOptions([]ClientOption{})

	assert.DeepEqual(t, consts.DefaultDialTimeout, options.DialTimeout)
	assert.DeepEqual(t, consts.DefaultMaxConnsPerHost, options.MaxConnsPerHost)
	assert.DeepEqual(t, consts.DefaultMaxIdleConnDuration, options.MaxIdleConnDuration)
	assert.DeepEqual(t, true, options.KeepAlive)
}

// TestCustomClientOptions test client options with custom values
func TestCustomClientOptions(t *testing.T) {
	options := NewClientOptions([]ClientOption{})

	options.Apply([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.DialTimeout = 2 * time.Second
			},
		},
	})
	assert.DeepEqual(t, 2*time.Second, options.DialTimeout)
}
