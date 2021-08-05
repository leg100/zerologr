package zerologr

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

// zeroLogger is a logr.LogSink that uses zerolog to log.
type zeroLogger struct {
	l *zerolog.Logger
}

type keyValue struct {
	key   string
	value interface{}
}

// Zerolog levels are int8 - make sure we stay in bounds.
func toZeroLevel(lvl int) zerolog.Level {
	if lvl > 127 {
		lvl = 127
	}
	// zerolog levels are both inverted and one higher (e.g. trace in zerolog is
	// -1 whereas in logr it is -2)
	return (0 - zerolog.Level(lvl)) + 1
}

func newKeyValues(args []interface{}) ([]keyValue, error) {
	if len(args) == 0 {
		return nil, nil
	}

	if len(args)%2 != 0 {
		return nil, errors.New("odd number of arguments passed as key-value pairs for logging")
	}

	kvs := make([]keyValue, 0, len(args)/2)

	for i := 0; i < len(args); i += 2 {
		k, v := args[i], args[i+1]
		kStr, isString := k.(string)
		if !isString {
			return nil, fmt.Errorf("non-string key argument passed to logging: %v", k)
		}
		kvs = append(kvs, keyValue{key: kStr, value: v})
	}

	return kvs, nil
}

// addToEvent adds key-value pairs to a zerolog event
func addToEvent(e *zerolog.Event, keysAndVals []keyValue) {
	for _, kv := range keysAndVals {
		e = e.Interface(kv.key, kv.value)
	}
}

// addToContext adds key-value pairs to a zerolog context
func addToContext(ctx zerolog.Context, keysAndVals []keyValue) zerolog.Context {
	for _, kv := range keysAndVals {
		ctx = ctx.Interface(kv.key, kv.value)
	}

	return ctx
}

// No-op
func (zl *zeroLogger) Init(ri logr.RuntimeInfo) {}

func (zl *zeroLogger) Enabled(lvl int) bool {
	return toZeroLevel(lvl) >= zl.l.GetLevel()
}

func (zl *zeroLogger) Info(lvl int, msg string, args ...interface{}) {
	kvs, err := newKeyValues(args)
	if err != nil {
		zl.l.Err(err).Msg("unable to log message")
		return
	}

	e := zl.l.WithLevel(toZeroLevel(lvl))
	addToEvent(e, kvs)
	e.Msg(msg)
}

func (zl *zeroLogger) Error(err error, msg string, args ...interface{}) {
	// Only log an error if level is low enough
	if zl.l.GetLevel() > zerolog.ErrorLevel {
		return
	}

	event := zl.l.Error().Err(err)

	kvs, err := newKeyValues(args)
	if err != nil {
		zl.l.Err(err).Msg("unable to log message")
		return
	}

	addToEvent(event, kvs)
	event.Msg(msg)
}

// WithName returns a new logger with the given name. Note: the name is
// currently ignored, as zerolog loggers do not have names.
func (zl *zeroLogger) WithName(_ string) logr.LogSink {
	newLogger := zl.l.With().Logger()
	return &zeroLogger{l: &newLogger}
}

func (zl *zeroLogger) WithValues(args ...interface{}) logr.LogSink {
	kvs, err := newKeyValues(args)
	if err != nil {
		zl.l.Err(err).Msg("unable to log message")
		return zl
	}

	ctx := addToContext(zl.l.With(), kvs)
	newLogger := ctx.Logger()
	return &zeroLogger{l: &newLogger}
}

// NewLogger creates a new logr.Logger using the given zerolog Logger to log.
func NewLogger(zl *zerolog.Logger) logr.Logger {
	return logr.New(&zeroLogger{l: zl})
}

var _ logr.LogSink = (*zeroLogger)(nil)
