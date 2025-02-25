package loadbalance

import (
	"math"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/client/discovery"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestWeightedBalancer(t *testing.T) {
	balancer := NewWeightedBalancer()
	// nil
	ins := balancer.Pick(discovery.Result{})
	assert.DeepEqual(t, ins, nil)

	// empty instance
	e := discovery.Result{
		Instances: make([]discovery.Instance, 0),
		CacheKey:  "a",
	}
	balancer.Rebalance(e)
	ins = balancer.Pick(e)
	assert.DeepEqual(t, ins, nil)

	// one instance
	insList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8888", 20, nil),
	}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "b",
	}
	balancer.Rebalance(e)
	for i := 0; i < 100; i++ {
		ins = balancer.Pick(e)
		assert.DeepEqual(t, ins.Weight(), 20)
	}

	// multi instances, weightSum > 0
	insList = []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 100, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", 200, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8883", 300, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8884", 400, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8885", 500, nil),
	}

	var weightSum int
	for _, ins := range insList {
		weight := ins.Weight()
		weightSum += weight
	}

	n := 1000000
	pickedStat := map[int]int{}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "c",
	}
	balancer.Rebalance(e)
	for i := 0; i < n; i++ {
		ins = balancer.Pick(e)
		weight := ins.Weight()
		if pickedCnt, ok := pickedStat[weight]; ok {
			pickedStat[weight] = pickedCnt + 1
		} else {
			pickedStat[weight] = 1
		}
	}

	for _, ins := range insList {
		weight := ins.Weight()
		expect := float64(weight) / float64(weightSum) * float64(n)
		actual := float64(pickedStat[weight])
		delta := math.Abs(expect - actual)
		assert.DeepEqual(t, true, delta/expect < 0.05)
	}

	// have instances that weight < 0
	insList = []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", -10, nil),
	}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "d",
	}
	balancer.Rebalance(e)
	for i := 0; i < 1000; i++ {
		ins = balancer.Pick(e)
		assert.DeepEqual(t, 10, ins.Weight())
	}
}
