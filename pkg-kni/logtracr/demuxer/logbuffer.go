package demuxer

import (
	"bytes"
	"fmt"
	"time"
)

const (
	LevelError = "ERROR"
	LevelInfo  = "INFO"
	LevelV     = "V[%d]"
)

func levelString(level int) string {
	if level > 0 {
		return fmt.Sprintf(LevelV, level)
	}
	return LevelInfo
}

type logBuffer struct {
	bytes.Buffer
	ts time.Time // last updated
}

func (lb *logBuffer) WriteLine(opts Options, name, level, msg string, values, kv []any) {
	lb.WriteString(lb.ts.Format(time.StampMilli))
	lb.WriteString(" ")
	lb.WriteString(level)
	if name != "" {
		lb.WriteString(" ")
		lb.WriteString(name)
	}
	if msg != "" {
		lb.WriteString(" ")
		lb.WriteString(msg)
	}
	if len(values) > 0 {
		lb.WriteString(" ")
		lb.WriteString(opts.KeyValueFormatter(values))
	}
	if len(kv) > 0 {
		lb.WriteString(" ")
		lb.WriteString(opts.KeyValueFormatter(kv))
	}
	lb.WriteString("\n")
}
