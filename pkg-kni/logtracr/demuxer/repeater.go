package demuxer

import (
	"time"

	"github.com/go-logr/logr"
)

type Repeater struct {
	opts    Options
	name    string
	values  []any
	backend logr.LogSink
}

func NewRepeater(backend logr.LogSink, opts Options) (*Repeater, error) {
	if opts.KeyFinder == nil {
		return nil, ErrMissingKeyFinder
	}
	if opts.KeyValueFormatter == nil {
		return nil, ErrMissingKeyValueFormatter
	}
	return &Repeater{
		opts:    opts,
		backend: backend,
	}, nil
}

// Init is not implemented and does not use any runtime info.
func (rp *Repeater) Init(info logr.RuntimeInfo) {
	// not implemented
}

// Enabled tests whether this Logger is enabled.
func (rp *Repeater) Enabled(level int) bool {
	// we filter using a different approach
	return matchLoggerByName(rp.name)
}

func (rp *Repeater) Info(level int, msg string, kv ...any) {
	rp.writeLine(levelString(level), msg, kv...)
}

func (rp *Repeater) Error(err error, msg string, kv ...any) {
	rp.writeLine(LevelError+" "+err.Error(), msg, kv...)
}

func (rp *Repeater) WithValues(kv ...any) logr.LogSink {
	return &Repeater{
		name:    rp.name,
		values:  append(rp.values, kv...),
		opts:    rp.opts,
		backend: rp.backend,
	}
}

func (rp *Repeater) WithName(name string) logr.LogSink {
	if rp.name != "" {
		name = rp.name + NameSeparator + name
	}
	return &Repeater{
		name:    name,
		values:  rp.values,
		opts:    rp.opts,
		backend: rp.backend,
	}
}

func (rp *Repeater) writeLine(level, msg string, kv ...any) {
	if len(kv) < 2 || len(kv)%2 != 0 {
		return
	}
	if rp.opts.KeyFinder == nil {
		return
	}

	_, ok := rp.opts.KeyFinder(kv)
	if !ok {
		return
	}

	ts := time.Now()
	logBuf := &logBuffer{}
	logBuf.ts = ts
	logBuf.WriteLine(rp.opts, rp.name, level, ">>> "+msg, rp.values, kv)

	rp.backend.Info(0, logBuf.String())
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &Repeater{}

// TODO: implement logr.CallDepthLogSink
