package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/telecom-cloud/client-go/pkg/client/discovery"
	"github.com/telecom-cloud/client-go/pkg/client/loadbalance"
	"github.com/telecom-cloud/client-go/pkg/common/registry"
)

// ServiceDiscoveryOptions service discovery option for client
type ServiceDiscoveryOptions struct {
	// Resolver is used to client discovery
	Resolver discovery.Resolver

	// Balancer is used to client load balance
	Balancer loadbalance.Loadbalancer

	// LbOpts LoadBalance option
	LbOpts loadbalance.Options
}

func (o *ServiceDiscoveryOptions) Apply(opts []ServiceDiscoveryOption) {
	for _, op := range opts {
		op.F(o)
	}
}

type ServiceDiscoveryOption struct {
	F func(o *ServiceDiscoveryOptions)
}

// WithCustomizedAddrs specifies the target instance addresses when doing service discovery.
// It overwrites the results from the Resolver
func WithCustomizedAddrs(addrs ...string) ServiceDiscoveryOption {
	return ServiceDiscoveryOption{
		F: func(o *ServiceDiscoveryOptions) {
			var ins []discovery.Instance
			for _, addr := range addrs {
				if _, err := net.ResolveTCPAddr("tcp", addr); err == nil {
					ins = append(ins, discovery.NewInstance("tcp", addr, registry.DefaultWeight, nil))
					continue
				}
				if _, err := net.ResolveUnixAddr("unix", addr); err == nil {
					ins = append(ins, discovery.NewInstance("unix", addr, registry.DefaultWeight, nil))
					continue
				}
				panic(fmt.Errorf("WithCustomizedAddrs: invalid '%s'", addr))
			}
			if len(ins) == 0 {
				panic("WithCustomizedAddrs() requires at least one argument")
			}

			targets := strings.Join(addrs, ",")
			o.Resolver = &discovery.SynthesizedResolver{
				ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
					return discovery.Result{
						CacheKey:  "fixed",
						Instances: ins,
					}, nil
				},
				NameFunc: func() string { return targets },
				TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
					return targets
				},
			}
		},
	}
}

// WithLoadBalanceOptions  sets Loadbalancer and loadbalance options for hertz client
func WithLoadBalanceOptions(lb loadbalance.Loadbalancer, options loadbalance.Options) ServiceDiscoveryOption {
	return ServiceDiscoveryOption{F: func(o *ServiceDiscoveryOptions) {
		o.LbOpts = options
		o.Balancer = lb
	}}
}
