package logger

import (
	"bytes"
	"testing"
)

func TestError_WritesToStderrWithTrailingNewline(t *testing.T) {
	var out, errBuf bytes.Buffer
	l := New(&out, &errBuf)

	l.Error("something broke")

	if errBuf.String() != "something broke\n" {
		t.Errorf("err: got %q", errBuf.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", out.String())
	}
}

func TestLog_WritesToStdoutWithTrailingNewline(t *testing.T) {
	var out bytes.Buffer
	l := New(&out, nil)

	l.Log("hello world")

	got := out.String()
	want := "hello world\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
