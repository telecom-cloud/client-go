package stats

import (
	"sync"
	"sync/atomic"

	"github.com/telecom-cloud/client-go/pkg/common/errors"
)

// EventIndex indicates a unique event.
type EventIndex int

// Level sets the record level.
type Level int

// Event levels.
const (
	LevelDisabled Level = iota
	LevelBase
	LevelDetailed
)

// Event is used to indicate a specific event.
type Event interface {
	Index() EventIndex
	Level() Level
}

type event struct {
	idx   EventIndex
	level Level
}

// Index implements the Event interface.
func (e event) Index() EventIndex {
	return e.idx
}

// Level implements the Event interface.
func (e event) Level() Level {
	return e.level
}

const (
	_ EventIndex = iota
	serverHandleStart
	serverHandleFinish
	httpStart
	httpFinish
	readHeaderStart
	readHeaderFinish
	readBodyStart
	readBodyFinish
	writeStart
	writeFinish
	predefinedEventNum
)

// Predefined events.
var (
	HTTPStart  = newEvent(httpStart, LevelBase)
	HTTPFinish = newEvent(httpFinish, LevelBase)

	ServerHandleStart  = newEvent(serverHandleStart, LevelDetailed)
	ServerHandleFinish = newEvent(serverHandleFinish, LevelDetailed)
	ReadHeaderStart    = newEvent(readHeaderStart, LevelDetailed)
	ReadHeaderFinish   = newEvent(readHeaderFinish, LevelDetailed)
	ReadBodyStart      = newEvent(readBodyStart, LevelDetailed)
	ReadBodyFinish     = newEvent(readBodyFinish, LevelDetailed)
	WriteStart         = newEvent(writeStart, LevelDetailed)
	WriteFinish        = newEvent(writeFinish, LevelDetailed)
)

// errors
var (
	ErrNotAllowed = errors.NewPublic("event definition is not allowed after initialization")
	ErrDuplicated = errors.NewPublic("event name is already defined")
)

var (
	lock        sync.RWMutex
	inited      int32
	userDefined = make(map[string]Event)
	maxEventNum = int(predefinedEventNum)
)

// FinishInitialization freezes all events defined and prevents further definitions to be added.
func FinishInitialization() {
	atomic.StoreInt32(&inited, 1)
}

// DefineNewEvent allows user to add event definitions during program initialization.
func DefineNewEvent(name string, level Level) (Event, error) {
	if atomic.LoadInt32(&inited) == 1 {
		return nil, ErrNotAllowed
	}
	lock.Lock()
	defer lock.Unlock()
	evt, exist := userDefined[name]
	if exist {
		return evt, ErrDuplicated
	}
	userDefined[name] = newEvent(EventIndex(maxEventNum), level)
	maxEventNum++
	return userDefined[name], nil
}

// MaxEventNum returns the number of event defined.
func MaxEventNum() int {
	lock.RLock()
	defer lock.RUnlock()
	return maxEventNum
}

func newEvent(idx EventIndex, level Level) Event {
	return event{
		idx:   idx,
		level: level,
	}
}
