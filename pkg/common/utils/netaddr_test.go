package utils

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestNetAddr(t *testing.T) {
	networkAddr := NewNetAddr("127.0.0.1", "192.168.1.1")

	assert.DeepEqual(t, networkAddr.Network(), "127.0.0.1")
	assert.DeepEqual(t, networkAddr.String(), "192.168.1.1")
}
