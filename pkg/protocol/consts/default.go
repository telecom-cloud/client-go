package consts

import "time"

const (
	// *** Server default value ***

	// DefaultMaxInMemoryFileSize defines the in memory file size when parse
	// multipart_form. If the size exceeds, then hertz will write to disk.
	DefaultMaxInMemoryFileSize = 16 * 1024 * 1024

	// *** Client default value start from here ***

	// DefaultDialTimeout is timeout used by Dialer and DialDualStack
	// for establishing TCP connections.
	DefaultDialTimeout = time.Second

	// DefaultMaxConnsPerHost is the maximum number of concurrent connections
	// http client may establish per host by default (i.e. if
	// Client.MaxConnsPerHost isn't set).
	DefaultMaxConnsPerHost = 512

	// DefaultMaxIdleConnDuration is the default duration before idle keep-alive
	// connection is closed.
	DefaultMaxIdleConnDuration = 10 * time.Second

	// DefaultMaxIdempotentCallAttempts is the default idempotent calls attempts count.
	DefaultMaxIdempotentCallAttempts = 1

	// DefaultMaxRetryTimes is the default call times of retry
	DefaultMaxRetryTimes = 1
)
