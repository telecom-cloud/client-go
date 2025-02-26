package stats

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestDefineNewEvent(t *testing.T) {
	num0 := MaxEventNum()

	event1, err1 := DefineNewEvent("myevent", LevelDetailed)
	num1 := MaxEventNum()

	assert.Assert(t, err1 == nil)
	assert.Assert(t, event1 != nil)
	assert.Assert(t, num1 == num0+1)
	assert.Assert(t, event1.Level() == LevelDetailed)

	event2, err2 := DefineNewEvent("myevent", LevelBase)
	num2 := MaxEventNum()
	assert.Assert(t, err2 == ErrDuplicated)
	assert.Assert(t, event2 == event1)
	assert.Assert(t, num2 == num1)
	assert.Assert(t, event2.Level() == LevelDetailed)

	FinishInitialization()

	event3, err3 := DefineNewEvent("another", LevelDetailed)
	num3 := MaxEventNum()
	assert.Assert(t, err3 == ErrNotAllowed)
	assert.Assert(t, event3 == nil)
	assert.Assert(t, num3 == num1)
}
