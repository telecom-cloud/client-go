package retry

import (
	"math"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestApply(t *testing.T) {
	delayPolicyFunc := func(attempts uint, err error, retryConfig *Config) time.Duration {
		return time.Second
	}
	options := []Option{}
	options = append(options, WithMaxAttemptTimes(100), WithInitDelay(time.Second),
		WithMaxDelay(time.Second), WithDelayPolicy(delayPolicyFunc), WithMaxJitter(time.Second))

	config := Config{}
	config.Apply(options)

	assert.DeepEqual(t, uint(100), config.MaxAttemptTimes)
	assert.DeepEqual(t, time.Second, config.Delay)
	assert.DeepEqual(t, time.Second, config.MaxDelay)
	assert.DeepEqual(t, time.Second, Delay(0, nil, &config))
	assert.DeepEqual(t, time.Second, config.MaxJitter)
}

func TestPolicy(t *testing.T) {
	dur := DefaultDelayPolicy(0, nil, nil)
	assert.DeepEqual(t, 0*time.Millisecond, dur)

	config := Config{
		Delay: time.Second,
	}
	dur = FixedDelayPolicy(0, nil, &config)
	assert.DeepEqual(t, time.Second, dur)

	dur = RandomDelayPolicy(0, nil, &config)
	assert.DeepEqual(t, 0*time.Millisecond, dur)
	config.MaxJitter = time.Second * 1
	dur = RandomDelayPolicy(0, nil, &config)
	assert.NotEqual(t, time.Second*1, dur)

	dur = BackOffDelayPolicy(0, nil, &config)
	assert.DeepEqual(t, time.Second*1, dur)
	config.Delay = time.Duration(-1)
	dur = BackOffDelayPolicy(0, nil, &config)
	assert.DeepEqual(t, time.Second*0, dur)
	config.Delay = time.Duration(1)
	dur = BackOffDelayPolicy(63, nil, &config)
	durExp := config.Delay << 62
	assert.DeepEqual(t, durExp, dur)

	dur = Delay(0, nil, &config)
	assert.DeepEqual(t, 0*time.Millisecond, dur)
	delayPolicyFunc := func(attempts uint, err error, retryConfig *Config) time.Duration {
		return time.Second
	}
	config.DelayPolicy = delayPolicyFunc
	config.MaxDelay = time.Second / 2
	dur = Delay(0, nil, &config)
	assert.DeepEqual(t, config.MaxDelay, dur)

	delayPolicyFunc2 := func(attempts uint, err error, retryConfig *Config) time.Duration {
		return time.Duration(math.MaxInt64)
	}
	delayFunc := CombineDelay(delayPolicyFunc2, delayPolicyFunc)
	dur = delayFunc(0, nil, &config)
	assert.DeepEqual(t, time.Duration(math.MaxInt64), dur)
}
