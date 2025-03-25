package fanout

import (
	"bytes"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/tonglil/buflogr"
)

func TestFanoutPassthrough(t *testing.T) {
	buf := bytes.Buffer{}
	log := buflogr.NewWithBuffer(&buf)

	tlog := logr.New(NewWithLeaves(log.GetSink()))
	tlog.Info("info message", "foo", 42)
	tlog.Error(errors.New("fake error"), "error message", "bar", 21)
	tlog.V(5).Info("info v=5 message", "fizz", "buzz")

	got := buf.String()
	exp := `INFO info message foo 42
ERROR fake error error message bar 21
V[5] info v=5 message fizz buzz
`
	if got != exp {
		t.Fatalf("got={%s} exp={%s}", got, exp)
	}
}

func TestFanoutDuplicateInterleaved(t *testing.T) {
	buf := bytes.Buffer{}
	log := buflogr.NewWithBuffer(&buf)

	tlog := logr.New(NewWithLeaves(log.GetSink(), log.GetSink()))
	tlog.Info("info message", "foo", 42)
	tlog.Error(errors.New("fake error"), "error message", "bar", 21)
	tlog.V(5).Info("info v=5 message", "fizz", "buzz")

	got := buf.String()
	exp := `INFO info message foo 42
INFO info message foo 42
ERROR fake error error message bar 21
ERROR fake error error message bar 21
V[5] info v=5 message fizz buzz
V[5] info v=5 message fizz buzz
`
	if got != exp {
		t.Fatalf("got={%s} exp={%s}", got, exp)
	}
}

func TestFanoutDuplicateDistinct(t *testing.T) {
	buf1 := bytes.Buffer{}
	log1 := buflogr.NewWithBuffer(&buf1)
	buf2 := bytes.Buffer{}
	log2 := buflogr.NewWithBuffer(&buf2)

	tlog := logr.New(NewWithLeaves(log1.GetSink(), log2.GetSink()))
	tlog.Info("info message", "foo", 42)
	tlog.Error(errors.New("fake error"), "error message", "bar", 21)
	tlog.V(5).Info("info v=5 message", "fizz", "buzz")

	got1 := buf1.String()
	got2 := buf2.String()
	exp := `INFO info message foo 42
ERROR fake error error message bar 21
V[5] info v=5 message fizz buzz
`
	if got1 != exp {
		t.Fatalf("got={%s} exp={%s}", got1, exp)
	}
	if got2 != exp {
		t.Fatalf("got={%s} exp={%s}", got2, exp)
	}
}

func TestFanoutDuplicateDistinctWithNameAndValues(t *testing.T) {
	buf1 := bytes.Buffer{}
	log1 := buflogr.NewWithBuffer(&buf1)
	buf2 := bytes.Buffer{}
	log2 := buflogr.NewWithBuffer(&buf2)

	tlog := logr.New(NewWithLeaves(log1.GetSink(), log2.GetSink()))

	tlog = tlog.WithName("foo").WithValues("abc", 123)
	tlog = tlog.WithName("bar")
	tlog.Info("info message", "foo", 42)

	tlog = tlog.WithName("baz")
	tlog = tlog.WithValues("A", 1, "B", 2)
	tlog.Error(errors.New("fake error"), "error message", "bar", 21)
	tlog.V(5).Info("info v=5 message", "fizz", "buzz")

	got1 := buf1.String()
	got2 := buf2.String()
	exp := `INFO foo/bar info message abc 123 foo 42
ERROR fake error foo/bar/baz error message abc 123 A 1 B 2 bar 21
V[5] foo/bar/baz info v=5 message abc 123 A 1 B 2 fizz buzz
`
	if got1 != exp {
		t.Fatalf("got={%s} exp={%s}", got1, exp)
	}
	if got2 != exp {
		t.Fatalf("got={%s} exp={%s}", got2, exp)
	}
}
