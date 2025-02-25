package traceinfo

import (
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/tracer/stats"
)

// HTTPStats is used to collect statistics about the HTTP.
type HTTPStats interface {
	Record(event stats.Event, status stats.Status, info string)
	GetEvent(event stats.Event) Event
	SendSize() int
	SetSendSize(size int)
	RecvSize() int
	SetRecvSize(size int)
	Error() error
	SetError(err error)
	Panicked() (bool, interface{})
	SetPanicked(x interface{})
	Level() stats.Level
	SetLevel(level stats.Level)
	Reset()
}

// Event is the abstraction of an event happened at a specific time.
type Event interface {
	Event() stats.Event
	Status() stats.Status
	Info() string
	Time() time.Time
	IsNil() bool
}

// TraceInfo contains the trace message in Crafter.
type TraceInfo interface {
	Stats() HTTPStats
	Reset()
}
