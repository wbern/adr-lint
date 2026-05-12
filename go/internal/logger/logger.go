// Package logger wraps stdout/stderr writes behind a Logger struct
// with injectable writers, so callers (and tests) can capture output
// without redirecting global state.
package logger

import (
	"fmt"
	"io"
	"os"
)

// Logger writes informational messages to Out and errors to Err.
type Logger struct {
	Out io.Writer
	Err io.Writer
}

// New builds a Logger. nil writers fall back to os.Stdout / os.Stderr.
func New(out, err io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return &Logger{Out: out, Err: err}
}

// Default is a process-wide Logger that writes to os.Stdout / os.Stderr.
var Default = New(nil, nil)

// Log writes message followed by a newline to Out.
func (l *Logger) Log(message string) {
	fmt.Fprintln(l.Out, message)
}

// Error writes args (space-separated, newline-terminated) to Err.
func (l *Logger) Error(args ...any) {
	fmt.Fprintln(l.Err, args...)
}
