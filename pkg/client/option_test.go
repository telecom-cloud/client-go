package client

import (
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/client/retry"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestClientOptions(t *testing.T) {
	opt := config.NewClientOptions([]config.ClientOption{
		WithDialTimeout(100 * time.Millisecond),
		WithMaxConnsPerHost(128),
		WithMaxIdleConnDuration(5 * time.Second),
		WithMaxConnDuration(10 * time.Second),
		WithMaxConnWaitTimeout(5 * time.Second),
		WithKeepAlive(false),
		WithClientReadTimeout(1 * time.Second),
		WithResponseBodyStream(true),
		WithRetryConfig(
			retry.WithMaxAttemptTimes(2),
			retry.WithInitDelay(100*time.Millisecond),
			retry.WithMaxDelay(5*time.Second),
			retry.WithMaxJitter(1*time.Second),
			retry.WithDelayPolicy(retry.CombineDelay(retry.DefaultDelayPolicy, retry.FixedDelayPolicy, retry.BackOffDelayPolicy)),
		),
		WithWriteTimeout(time.Second),
		WithConnStateObserve(nil, time.Second),
	})
	assert.DeepEqual(t, 100*time.Millisecond, opt.DialTimeout)
	assert.DeepEqual(t, 128, opt.MaxConnsPerHost)
	assert.DeepEqual(t, 5*time.Second, opt.MaxIdleConnDuration)
	assert.DeepEqual(t, 10*time.Second, opt.MaxConnDuration)
	assert.DeepEqual(t, 5*time.Second, opt.MaxConnWaitTimeout)
	assert.DeepEqual(t, false, opt.KeepAlive)
	assert.DeepEqual(t, 1*time.Second, opt.ReadTimeout)
	assert.DeepEqual(t, 1*time.Second, opt.WriteTimeout)
	assert.DeepEqual(t, true, opt.ResponseBodyStream)
	assert.DeepEqual(t, uint(2), opt.RetryConfig.MaxAttemptTimes)
	assert.DeepEqual(t, 100*time.Millisecond, opt.RetryConfig.Delay)
	assert.DeepEqual(t, 5*time.Second, opt.RetryConfig.MaxDelay)
	assert.DeepEqual(t, 1*time.Second, opt.RetryConfig.MaxJitter)
	assert.DeepEqual(t, 1*time.Second, opt.ObservationInterval)
	for i := 0; i < 100; i++ {
		assert.DeepEqual(t, opt.RetryConfig.DelayPolicy(uint(i), nil, opt.RetryConfig), retry.CombineDelay(retry.DefaultDelayPolicy, retry.FixedDelayPolicy, retry.BackOffDelayPolicy)(uint(i), nil, opt.RetryConfig))
	}
}
