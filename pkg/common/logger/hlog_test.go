package logger

import (
	"log"
	"os"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestDefaultAndSysLogger(t *testing.T) {
	defaultLog := DefaultLogger()
	systemLog := SystemLogger()

	assert.DeepEqual(t, logger, defaultLog)
	assert.DeepEqual(t, sysLogger, systemLog)
	assert.NotEqual(t, logger, systemLog)
	assert.NotEqual(t, sysLogger, defaultLog)
}

func TestSetLogger(t *testing.T) {
	setLog := &defaultLogger{
		stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
		depth:  6,
	}
	setSysLog := &systemLogger{
		setLog,
		systemLogPrefix,
	}

	assert.NotEqual(t, logger, setLog)
	assert.NotEqual(t, sysLogger, setSysLog)
	SetLogger(setLog)
	assert.DeepEqual(t, logger, setLog)
	assert.DeepEqual(t, sysLogger, setSysLog)
}
