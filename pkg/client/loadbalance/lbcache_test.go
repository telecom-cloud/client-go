package loadbalance

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/client/discovery"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

func TestBuilder(t *testing.T) {
	ins := discovery.NewInstance("tcp", "127.0.0.1:8888", 10, map[string]string{"a": "b"})
	r := &discovery.SynthesizedResolver{
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{CacheKey: key, Instances: []discovery.Instance{ins}}, nil
		},
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return "mockRoute"
		},
		NameFunc: func() string { return t.Name() },
	}
	lb := mockLoadbalancer{
		rebalanceFunc: nil,
		deleteFunc:    nil,
		pickFunc: func(res discovery.Result) discovery.Instance {
			assert.Assert(t, res.CacheKey == t.Name()+":mockRoute", res.CacheKey)
			assert.Assert(t, len(res.Instances) == 1)
			assert.Assert(t, len(res.Instances) == 1)
			assert.Assert(t, res.Instances[0].Address().String() == "127.0.0.1:8888")
			return res.Instances[0]
		},
		nameFunc: func() string { return "Synthesized" },
	}
	NewBalancerFactory(Config{
		Balancer: lb,
		LbOpts:   DefaultLbOpts,
		Resolver: r,
	})
	b, ok := balancerFactories.Load(cacheKey(t.Name(), "Synthesized", DefaultLbOpts))
	assert.Assert(t, ok)
	assert.Assert(t, b != nil)
	req := &protocol.Request{}
	req.SetHost("hertz.api.test")
	ins1, err := b.(*BalancerFactory).GetInstance(context.TODO(), req)
	assert.Assert(t, err == nil)
	assert.Assert(t, ins1.Address().String() == "127.0.0.1:8888")
	assert.Assert(t, ins1.Weight() == 10)
	value, exists := ins1.Tag("a")
	assert.Assert(t, value == "b")
	assert.Assert(t, exists == true)
}

func TestBalancerCache(t *testing.T) {
	count := 10
	inss := make([]discovery.Instance, 0, count)
	for i := 0; i < count; i++ {
		inss = append(inss, discovery.NewInstance("tcp", fmt.Sprint(i), 10, nil))
	}
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return target.Host
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{CacheKey: "svc", Instances: inss}, nil
		},
		NameFunc: func() string { return t.Name() },
	}
	lb := NewWeightedBalancer()
	for i := 0; i < count; i++ {
		blf := NewBalancerFactory(Config{
			Balancer: lb,
			LbOpts:   Options{},
			Resolver: r,
		})
		req := &protocol.Request{}
		req.SetHost("svc")
		for a := 0; a < count; a++ {
			addr, err := blf.GetInstance(context.TODO(), req)
			assert.Assert(t, err == nil, err)
			t.Logf("count: %d addr: %s\n", i, addr.Address().String())
		}
	}
}

func TestBalancerRefresh(t *testing.T) {
	var ins atomic.Value
	ins.Store(discovery.NewInstance("tcp", "127.0.0.1:8888", 10, nil))
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return target.Host
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{CacheKey: "svc1", Instances: []discovery.Instance{ins.Load().(discovery.Instance)}}, nil
		},
		NameFunc: func() string { return t.Name() },
	}
	opts := DefaultLbOpts
	opts.RefreshInterval = 30 * time.Millisecond
	blf := NewBalancerFactory(Config{
		Balancer: NewWeightedBalancer(),
		LbOpts:   opts,
		Resolver: r,
	})
	req := &protocol.Request{}
	req.SetHost("svc1")
	addr, err := blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	assert.Assert(t, addr.Address().String() == "127.0.0.1:8888")
	ins.Store(discovery.NewInstance("tcp", "127.0.0.1:8889", 10, nil))
	addr, err = blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	assert.Assert(t, addr.Address().String() == "127.0.0.1:8888")
	time.Sleep(2 * opts.RefreshInterval)
	addr, err = blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	assert.Assert(t, addr.Address().String() == "127.0.0.1:8889")
}

func TestBalancerExpires(t *testing.T) {
	n := int32(1000)
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return target.Host
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			ins := discovery.NewInstance("tcp", "127.0.0.1:"+strconv.Itoa(int(atomic.AddInt32(&n, 1))), 10, nil)
			return discovery.Result{CacheKey: "svc1", Instances: []discovery.Instance{ins}}, nil
		},
		NameFunc: func() string { return t.Name() },
	}
	opts := DefaultLbOpts
	opts.ExpireInterval = 30 * time.Millisecond
	blf := NewBalancerFactory(Config{
		Balancer: NewWeightedBalancer(),
		LbOpts:   opts,
		Resolver: r,
	})
	req := &protocol.Request{}
	req.SetHost("svc1")
	addr1, err := blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	addr2, err := blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	assert.Assert(t, addr1.Address().String() == addr2.Address().String())
	time.Sleep(3 * opts.ExpireInterval)
	addr3, err := blf.GetInstance(context.Background(), req)
	assert.Assert(t, err == nil, err)
	assert.Assert(t, addr3.Address().String() != addr2.Address().String())
}

func TestCacheKey(t *testing.T) {
	uniqueKey := cacheKey("hello", "world", Options{RefreshInterval: 15 * time.Second, ExpireInterval: 5 * time.Minute})
	assert.Assert(t, uniqueKey == "hello|world|{15s 5m0s}")
}

type mockLoadbalancer struct {
	rebalanceFunc func(ch discovery.Result)
	deleteFunc    func(key string)
	pickFunc      func(discovery.Result) discovery.Instance
	nameFunc      func() string
}

// Rebalance implements the Loadbalancer interface.
func (m mockLoadbalancer) Rebalance(ch discovery.Result) {
	if m.rebalanceFunc != nil {
		m.rebalanceFunc(ch)
	}
}

// Delete implements the Loadbalancer interface.
func (m mockLoadbalancer) Delete(ch string) {
	if m.deleteFunc != nil {
		m.deleteFunc(ch)
	}
}

// Name implements the Loadbalancer interface.
func (m mockLoadbalancer) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return ""
}

// Pick implements the Loadbalancer interface.
func (m mockLoadbalancer) Pick(d discovery.Result) discovery.Instance {
	if m.pickFunc != nil {
		return m.pickFunc(d)
	}
	return nil
}
