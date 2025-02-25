package discovery

import (
	"context"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/client/discovery"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

func TestDiscovery(t *testing.T) {
	inss := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8888", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8889", 10, nil),
	}
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return target.Host
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{CacheKey: "svc1", Instances: inss}, nil
		},
		NameFunc: func() string { return t.Name() },
	}

	mw := Discovery(r)
	checkMdw := func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		t.Log(string(req.Host()))
		assert.Assert(t, string(req.Host()) == "127.0.0.1:8888" || string(req.Host()) == "127.0.0.1:8889")
		return nil
	}
	for i := 0; i < 10; i++ {
		req := &protocol.Request{}
		resp := &protocol.Response{}
		req.Options().Apply([]config.RequestOption{config.WithSD(true)})
		req.SetRequestURI("http://service_name")
		_ = mw(checkMdw)(context.Background(), req, resp)
	}
}
