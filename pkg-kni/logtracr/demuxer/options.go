package demuxer

import (
	"fmt"
	"strings"
)

type Options struct {
	// KeyFinder can assume len(kv) >= 2 && len(kv)%2 == 0
	KeyFinder func(kv []any) (string, bool)
	// KeyValueFormatter can assume len(kv) > 0
	KeyValueFormatter func(kv []any) string
}

func DefaultKeyValueFormatter(kv []any) string {
	var sb strings.Builder
	if s, ok := toString(kv[0]); ok {
		sb.WriteString(s)
	}
	for _, x := range kv[1:] {
		if s, ok := toString(x); ok {
			sb.WriteString(" ")
			sb.WriteString(s)
		}
	}
	return sb.String()
}

func DefaultKeyFinder(key string) func(kv []any) (string, bool) {
	return func(kv []any) (string, bool) {
		if s, ok := toString(kv[0]); !ok || s != key {
			return "", false
		}
		return toString(kv[1])
	}
}

func GenericKeyFinder(key string) func(kv []any) (string, bool) {
	return func(kv []any) (string, bool) {
		for idx := 0; idx < len(kv); idx += 2 {
			if s, ok := toString(kv[idx]); ok && s == key {
				return toString(kv[idx+1])
			}
		}
		return "", false
	}
}

func toString(v any) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	if st, ok := v.(fmt.Stringer); ok {
		return st.String(), true
	}
	return "<unrep>", false
}
