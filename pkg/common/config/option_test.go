package config

import (
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/registry"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

// TestDefaultOptions test options with default values
func TestDefaultOptions(t *testing.T) {
	options := NewOptions([]Option{})

	assert.DeepEqual(t, defaultKeepAliveTimeout, options.KeepAliveTimeout)
	assert.DeepEqual(t, defaultReadTimeout, options.ReadTimeout)
	assert.DeepEqual(t, defaultReadTimeout, options.IdleTimeout)
	assert.DeepEqual(t, time.Duration(0), options.WriteTimeout)
	assert.True(t, options.RedirectTrailingSlash)
	assert.True(t, options.RedirectTrailingSlash)
	assert.False(t, options.HandleMethodNotAllowed)
	assert.False(t, options.UseRawPath)
	assert.False(t, options.RemoveExtraSlash)
	assert.True(t, options.UnescapePathValues)
	assert.False(t, options.DisablePreParseMultipartForm)
	assert.False(t, options.SenseClientDisconnection)
	assert.DeepEqual(t, defaultNetwork, options.Network)
	assert.DeepEqual(t, defaultAddr, options.Addr)
	assert.DeepEqual(t, defaultMaxRequestBodySize, options.MaxRequestBodySize)
	assert.False(t, options.GetOnly)
	assert.False(t, options.DisableKeepalive)
	assert.False(t, options.NoDefaultServerHeader)
	assert.DeepEqual(t, defaultWaitExitTimeout, options.ExitWaitTimeout)
	assert.Nil(t, options.TLS)
	assert.DeepEqual(t, defaultReadBufferSize, options.ReadBufferSize)
	assert.False(t, options.ALPN)
	assert.False(t, options.H2C)
	assert.DeepEqual(t, []interface{}{}, options.Tracers)
	assert.DeepEqual(t, new(interface{}), options.TraceLevel)
	assert.DeepEqual(t, registry.NoopRegistry, options.Registry)
	assert.Nil(t, options.BindConfig)
	assert.Nil(t, options.ValidateConfig)
	assert.Nil(t, options.CustomBinder)
	assert.Nil(t, options.CustomValidator)
	assert.DeepEqual(t, false, options.DisableHeaderNamesNormalizing)
}

// TestApplyCustomOptions test apply options with custom values after init
func TestApplyCustomOptions(t *testing.T) {
	options := NewOptions([]Option{})
	options.Apply([]Option{
		{F: func(o *Options) {
			o.Network = "unix"
		}},
	})
	assert.DeepEqual(t, "unix", options.Network)
}
