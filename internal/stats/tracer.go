package stats

import (
	"context"
	"runtime/debug"

	"github.com/telecom-cloud/client-go/pkg/app"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/common/tracer"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/stats"
)

// Controller controls tracers.
type Controller struct {
	tracers []tracer.Tracer
}

// Append appends a new tracer to the controller.
func (ctl *Controller) Append(col tracer.Tracer) {
	ctl.tracers = append(ctl.tracers, col)
}

// DoStart starts the tracers.
func (ctl *Controller) DoStart(ctx context.Context, c *app.RequestContext) context.Context {
	defer ctl.tryRecover()
	Record(c.GetTraceInfo(), stats.HTTPStart, nil)

	for _, col := range ctl.tracers {
		ctx = col.Start(ctx, c)
	}
	return ctx
}

// DoFinish calls the tracers in reversed order.
func (ctl *Controller) DoFinish(ctx context.Context, c *app.RequestContext, err error) {
	defer ctl.tryRecover()
	Record(c.GetTraceInfo(), stats.HTTPFinish, err)
	if err != nil {
		c.GetTraceInfo().Stats().SetError(err)
	}

	// reverse the order
	for i := len(ctl.tracers) - 1; i >= 0; i-- {
		ctl.tracers[i].Finish(ctx, c)
	}
}

func (ctl *Controller) HasTracer() bool {
	return ctl != nil && len(ctl.tracers) > 0
}

func (ctl *Controller) tryRecover() {
	if err := recover(); err != nil {
		logger.SystemLogger().Warnf("Panic happened during tracer call. This doesn't affect the http call, but may lead to lack of monitor data such as metrics and logs: %s, %s", err, string(debug.Stack()))
	}
}
