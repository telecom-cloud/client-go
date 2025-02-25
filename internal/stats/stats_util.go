package stats

import (
	"github.com/telecom-cloud/client-go/pkg/common/tracer/stats"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/traceinfo"
)

// Record records the event to HTTPStats.
func Record(ti traceinfo.TraceInfo, event stats.Event, err error) {
	if ti == nil {
		return
	}
	if err != nil {
		ti.Stats().Record(event, stats.StatusError, err.Error())
	} else {
		ti.Stats().Record(event, stats.StatusInfo, "")
	}
}

// CalcEventCostUs calculates the duration between start and end and returns in microsecond.
func CalcEventCostUs(start, end traceinfo.Event) uint64 {
	if start == nil || end == nil || start.IsNil() || end.IsNil() {
		return 0
	}
	return uint64(end.Time().Sub(start.Time()).Microseconds())
}
