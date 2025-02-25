package logger

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func initTestSysLogger() {
	sysLogger = &systemLogger{
		&defaultLogger{
			stdlog: log.New(os.Stderr, "", 0),
			depth:  4,
		},
		systemLogPrefix,
	}
}

func TestSysLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	sysLogger.Trace("trace work")
	sysLogger.Debug("received work order")
	sysLogger.Info("starting work")
	sysLogger.Notice("something happens in work")
	sysLogger.Warn("work may fail")
	sysLogger.Error("work failed")

	assert.DeepEqual(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}

func TestSysFormatLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	work := "work"
	sysLogger.Tracef("trace %s", work)
	sysLogger.Debugf("received %s order", work)
	sysLogger.Infof("starting %s", work)
	sysLogger.Noticef("something happens in %s", work)
	sysLogger.Warnf("%s may fail", work)
	sysLogger.Errorf("%s failed", work)

	assert.DeepEqual(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}

func TestSysCtxLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	ctx := context.Background()
	work := "work"
	sysLogger.CtxTracef(ctx, "trace %s", work)
	sysLogger.CtxDebugf(ctx, "received %s order", work)
	sysLogger.CtxInfof(ctx, "starting %s", work)
	sysLogger.CtxNoticef(ctx, "something happens in %s", work)
	sysLogger.CtxWarnf(ctx, "%s may fail", work)
	sysLogger.CtxErrorf(ctx, "%s failed", work)

	assert.DeepEqual(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}
