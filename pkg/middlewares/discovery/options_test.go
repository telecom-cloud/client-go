package discovery

import (
	"context"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/client/loadbalance"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestWithCustomizedAddrs(t *testing.T) {
	var options []ServiceDiscoveryOption
	options = append(options, WithCustomizedAddrs("127.0.0.1:8080", "/tmp/unix_ss"))
	opts := &ServiceDiscoveryOptions{}
	opts.Apply(options)
	assert.Assert(t, opts.Resolver.Name() == "127.0.0.1:8080,/tmp/unix_ss")
	res, err := opts.Resolver.Resolve(context.Background(), "")
	assert.Assert(t, err == nil)
	assert.Assert(t, res.Instances[0].Address().String() == "127.0.0.1:8080")
	assert.Assert(t, res.Instances[1].Address().String() == "/tmp/unix_ss")
}

func TestWithLoadBalanceOptions(t *testing.T) {
	balance := loadbalance.NewWeightedBalancer()
	var options []ServiceDiscoveryOption
	options = append(options, WithLoadBalanceOptions(balance, loadbalance.DefaultLbOpts))
	opts := &ServiceDiscoveryOptions{}
	opts.Apply(options)
	assert.Assert(t, opts.Balancer.Name() == "weight_random")
}
