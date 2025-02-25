package traceinfo

type traceInfo struct {
	stats HTTPStats
}

// Stats implements the HTTPInfo interface.
func (r *traceInfo) Stats() HTTPStats { return r.stats }

// Reset reuses the traceInfo.
func (r *traceInfo) Reset() {
	r.stats.Reset()
}

// NewTraceInfo creates a new traceInfoImpl using the given information.
func NewTraceInfo() TraceInfo {
	return &traceInfo{stats: NewHTTPStats()}
}
