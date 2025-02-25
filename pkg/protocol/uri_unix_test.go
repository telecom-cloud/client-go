//go:build !windows
// +build !windows

package protocol

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestGetScheme(t *testing.T) {
	scheme, path := getScheme([]byte("https://foo.com"))
	assert.DeepEqual(t, "https", string(scheme))
	assert.DeepEqual(t, "//foo.com", string(path))

	scheme, path = getScheme([]byte(":"))
	assert.DeepEqual(t, "", string(scheme))
	assert.DeepEqual(t, "", string(path))

	scheme, path = getScheme([]byte("ws://127.0.0.1"))
	assert.DeepEqual(t, "ws", string(scheme))
	assert.DeepEqual(t, "//127.0.0.1", string(path))

	scheme, path = getScheme([]byte("/hertz/demo"))
	assert.DeepEqual(t, "", string(scheme))
	assert.DeepEqual(t, "/hertz/demo", string(path))
}
