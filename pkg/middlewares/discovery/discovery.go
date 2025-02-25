package discovery

import (
	"context"

	"github.com/telecom-cloud/client-go/pkg/client"
	"github.com/telecom-cloud/client-go/pkg/client/discovery"
	"github.com/telecom-cloud/client-go/pkg/client/loadbalance"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// Discovery will construct a middleware with BalancerFactory.
func Discovery(resolver discovery.Resolver, opts ...ServiceDiscoveryOption) client.Middleware {
	options := &ServiceDiscoveryOptions{
		Balancer: loadbalance.NewWeightedBalancer(),
		LbOpts:   loadbalance.DefaultLbOpts,
		Resolver: resolver,
	}
	options.Apply(opts)

	lbConfig := loadbalance.Config{
		Resolver: options.Resolver,
		Balancer: options.Balancer,
		LbOpts:   options.LbOpts,
	}

	f := loadbalance.NewBalancerFactory(lbConfig)
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
			if req.Options() != nil && req.Options().IsSD() {
				ins, err := f.GetInstance(ctx, req)
				if err != nil {
					return err
				}
				req.SetHost(ins.Address().String())
			}
			return next(ctx, req, resp)
		}
	}
}
