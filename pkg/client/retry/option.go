package retry

import "time"

// Option is the only struct that can be used to set Retry Config.
type Option struct {
	F func(o *Config)
}

// WithMaxAttemptTimes set WithMaxAttemptTimes , including the first call.
func WithMaxAttemptTimes(maxAttemptTimes uint) Option {
	return Option{F: func(o *Config) {
		o.MaxAttemptTimes = maxAttemptTimes
	}}
}

// WithInitDelay set init Delay.
func WithInitDelay(delay time.Duration) Option {
	return Option{F: func(o *Config) {
		o.Delay = delay
	}}
}

// WithMaxDelay set MaxDelay.
func WithMaxDelay(maxDelay time.Duration) Option {
	return Option{F: func(o *Config) {
		o.MaxDelay = maxDelay
	}}
}

// WithDelayPolicy set DelayPolicy.
func WithDelayPolicy(delayPolicy DelayPolicyFunc) Option {
	return Option{F: func(o *Config) {
		o.DelayPolicy = delayPolicy
	}}
}

// WithMaxJitter set MaxJitter.
func WithMaxJitter(maxJitter time.Duration) Option {
	return Option{F: func(o *Config) {
		o.MaxJitter = maxJitter
	}}
}
