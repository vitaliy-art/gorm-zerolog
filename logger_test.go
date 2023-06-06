package gormzerolog

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

type testingEvent struct {
	added map[string]string
	msg   string
}

func (e *testingEvent) Str(key, value string) Event {
	if e.added == nil {
		e.added = map[string]string{}
	}

	e.added[key] = value
	return e
}

func (e *testingEvent) Msgf(format string, v ...any) {
	e.msg = fmt.Sprintf(format, v...)
}

func TestGormLogger(t *testing.T) {
	getEventFactory := func(e Event) func() Event {
		return func() Event { return e }
	}

	levelTest := func(logLevel logger.LogLevel) {
		assert := assert.New(t)
		assert.Containsf(
			[]logger.LogLevel{logger.Info, logger.Warn, logger.Error},
			logLevel,
			"unexpected logLevel value: %d",
			logLevel,
		)

		var (
			infoEvent   = &testingEvent{}
			warnEvent   = &testingEvent{}
			errorEvent  = &testingEvent{}
			targetEvent *testingEvent
			emptyEvents []*testingEvent
			l           = NewGormLogger().
					WithInfo(getEventFactory(infoEvent)).
					WithWarn(getEventFactory(warnEvent)).
					WithError(getEventFactory(errorEvent)).
					LogMode(logLevel).(*GormLogger)

			str1 = uuid.NewString()
			str2 = uuid.NewString()
			str3 = uuid.NewString()
			str4 = uuid.NewString()
			str5 = uuid.NewString()
			msg  = fmt.Sprintf("%s%%s", str4)
		)

		clearEvents := func() {
			infoEvent.added = nil
			infoEvent.msg = ""
			warnEvent.added = nil
			warnEvent.msg = ""
			errorEvent.added = nil
			errorEvent.msg = ""
		}

		assert.False(l.ignoreRecordNotFoundErr, "ignoreRecordNotFoundErr should be false")
		l.IgnoreRecordNotFoundError(true)
		assert.True(l.ignoreRecordNotFoundErr, "ignoreRecordNotFoundErr should be true")
		assert.Equal(l.slowThreshold, time.Millisecond*200)
		l.SlowThreshold(time.Millisecond * 600)
		assert.Equal(l.slowThreshold, time.Millisecond*600)
		l.AdditionalData = map[string]string{str1: str1, str2: str2, str3: str3}
		assert.Equalf(logLevel, l.logLevel, "logLevel should be %d", logLevel)
		switch logLevel {
		case logger.Info:
			l.Info(context.Background(), msg, str5)
			targetEvent = infoEvent
			emptyEvents = []*testingEvent{warnEvent, errorEvent}
		case logger.Warn:
			l.Warn(context.Background(), msg, str5)
			targetEvent = warnEvent
			emptyEvents = []*testingEvent{infoEvent, errorEvent}
		case logger.Error:
			l.Error(context.Background(), msg, str5)
			targetEvent = errorEvent
			emptyEvents = []*testingEvent{infoEvent, warnEvent}
		}

		assert.Equal(map[string]string{str1: str1, str2: str2, str3: str3}, targetEvent.added)
		assert.Equal(targetEvent.msg, fmt.Sprintf("%s%s", str4, str5))
		for _, e := range emptyEvents {
			assert.Empty(e.added)
			assert.Empty(e.msg)
		}

		clearEvents()
		l.LogMode(logger.Silent)
		switch logLevel {
		case logger.Info:
			l.Info(context.Background(), msg, str5)
		case logger.Warn:
			l.Warn(context.Background(), msg, str5)
		case logger.Error:
			l.Error(context.Background(), msg, str5)
		}

		for _, e := range append(emptyEvents, targetEvent) {
			assert.Empty(e.added)
			assert.Empty(e.msg)
		}

		l.Trace(context.Background(), time.Now(), func() (string, int64) { return "", 0 }, errors.New("test"))
		for _, e := range append(emptyEvents, targetEvent) {
			assert.Empty(e.added)
			assert.Empty(e.msg)
		}

		l.LogMode(logLevel)
		l.Trace(context.Background(), time.Now(), func() (string, int64) { return "test", 0 }, errors.New("test"))
		assert.NotEmpty(errorEvent.added)
		assert.NotEmpty(errorEvent.msg)
		assert.Empty(warnEvent.added)
		assert.Empty(warnEvent.msg)
		if logLevel >= logger.Info {
			assert.NotEmpty(infoEvent.added)
			assert.NotEmpty(infoEvent.msg)
		} else {
			assert.Empty(infoEvent.added)
			assert.Empty(infoEvent.msg)
		}

		clearEvents()
		l.Trace(context.Background(), time.Now().Add(-l.slowThreshold*2), func() (string, int64) { return "test", -1 }, nil)
		assert.Empty(errorEvent.added)
		assert.Empty(errorEvent.msg)
		if logLevel >= logger.Warn {
			assert.NotEmpty(warnEvent.added)
			assert.NotEmpty(warnEvent.msg)
		} else {
			assert.Empty(warnEvent.added)
			assert.Empty(warnEvent.msg)
		}
		if logLevel >= logger.Info {
			assert.NotEmpty(infoEvent.added)
			assert.NotEmpty(infoEvent.msg)
		} else {
			assert.Empty(infoEvent.added)
			assert.Empty(infoEvent.msg)
		}
	}

	t.Run("info test", func(t *testing.T) { levelTest(logger.Info) })
	t.Run("warn test", func(t *testing.T) { levelTest(logger.Warn) })
	t.Run("error test", func(t *testing.T) { levelTest(logger.Error) })
}
