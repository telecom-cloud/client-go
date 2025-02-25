package discovery

import (
	"context"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/registry"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestInstance(t *testing.T) {
	network := "192.168.1.1"
	address := "/hello"
	weight := 1
	instance := NewInstance(network, address, weight, nil)

	assert.DeepEqual(t, network, instance.Address().Network())
	assert.DeepEqual(t, address, instance.Address().String())
	assert.DeepEqual(t, weight, instance.Weight())
	val, ok := instance.Tag("name")
	assert.DeepEqual(t, "", val)
	assert.False(t, ok)

	instance2 := NewInstance("", "", 0, nil)
	assert.DeepEqual(t, registry.DefaultWeight, instance2.Weight())
}

func TestSynthesizedResolver(t *testing.T) {
	targetFunc := func(ctx context.Context, target *TargetInfo) string {
		return "hello"
	}
	resolveFunc := func(ctx context.Context, key string) (Result, error) {
		return Result{CacheKey: "name"}, nil
	}
	nameFunc := func() string {
		return "raymonder"
	}
	resolver := SynthesizedResolver{
		TargetFunc:  targetFunc,
		ResolveFunc: resolveFunc,
		NameFunc:    nameFunc,
	}

	assert.DeepEqual(t, "hello", resolver.Target(context.Background(), &TargetInfo{}))
	res, err := resolver.Resolve(context.Background(), "")
	assert.DeepEqual(t, "name", res.CacheKey)
	assert.Nil(t, err)
	assert.DeepEqual(t, "raymonder", resolver.Name())

	resolver2 := SynthesizedResolver{
		TargetFunc:  nil,
		ResolveFunc: nil,
		NameFunc:    nil,
	}
	assert.DeepEqual(t, "", resolver2.Target(context.Background(), &TargetInfo{}))
	assert.DeepEqual(t, "", resolver2.Name())
}
