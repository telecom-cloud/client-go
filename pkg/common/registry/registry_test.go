package registry

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestNoopRegistry(t *testing.T) {
	reg := noopRegistry{}
	assert.Nil(t, reg.Deregister(&Info{}))
	assert.Nil(t, reg.Register(&Info{}))
}
