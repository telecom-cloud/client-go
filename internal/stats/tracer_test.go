package stats

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/app"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/traceinfo"
)

type mockTracer struct {
	order         int
	stack         *[]int
	panicAtStart  bool
	panicAtFinish bool
}

func (mt *mockTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	if mt.panicAtStart {
		panic(fmt.Sprintf("panicked at start: Tracer(%d)", mt.order))
	}
	*mt.stack = append(*mt.stack, mt.order)
	return context.WithValue(ctx, mt, mt.order)
}

func (mt *mockTracer) Finish(ctx context.Context, c *app.RequestContext) {
	if mt.panicAtFinish {
		panic(fmt.Sprintf("panicked at finish: Tracer(%d)", mt.order))
	}
	*mt.stack = append(*mt.stack, -mt.order)
}

func TestOrder(t *testing.T) {
	var c Controller
	var stack []int
	t1 := &mockTracer{order: 1, stack: &stack}
	t2 := &mockTracer{order: 2, stack: &stack}
	ctx := app.NewContext(16)
	c.Append(t1)
	c.Append(t2)

	ctx0 := context.Background()
	ctx1 := c.DoStart(ctx0, ctx)
	assert.Assert(t, ctx1 != ctx0)
	assert.Assert(t, len(stack) == 2 && stack[0] == 1 && stack[1] == 2, stack)

	c.DoFinish(ctx1, ctx, nil)
	assert.Assert(t, len(stack) == 4 && stack[2] == -2 && stack[3] == -1, stack)
}

func TestPanic(t *testing.T) {
	var c Controller
	var stack []int
	t1 := &mockTracer{order: 1, stack: &stack, panicAtStart: true, panicAtFinish: true}
	t2 := &mockTracer{order: 2, stack: &stack}
	ctx := app.NewContext(16)
	ctx.SetTraceInfo(traceinfo.NewTraceInfo())
	c.Append(t1)
	c.Append(t2)

	ctx0 := context.Background()
	ctx1 := c.DoStart(ctx0, ctx)
	assert.Assert(t, ctx1 != ctx0)
	assert.Assert(t, len(stack) == 0) // t1's panic skips all subsequent Starts

	err := errors.New("some error")
	c.DoFinish(ctx1, ctx, err)
	assert.Assert(t, len(stack) == 1 && stack[0] == -2, stack)
}
