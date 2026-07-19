// Package stack captures source locations for errors without changing standard
// Go error-chain behavior.
package stack

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const maxFrames = 32

// Provider is an error that carries a captured source stack.
type Provider interface {
	error
	Stack() string
}

type stackError struct {
	err error
	pcs []uintptr

	once  sync.Once
	stack string
}

type replacementError struct {
	replacement error
	err         error
}

// With returns err wrapped with the caller's source stack.
func With(err error) error {
	return with(err, 0)
}

// WithSkip returns err wrapped with a source stack after skipping skip
// additional caller frames.
func WithSkip(err error, skip int) error {
	return with(err, skip)
}

// Unwrap removes one top-level stack wrapper.
func Unwrap(err error) error {
	if wrapped, ok := err.(*stackError); ok {
		return wrapped.err
	}
	return err
}

// Capture returns a source stack for fallback logging.
func Capture(skip int) string {
	return format(capturePCs(skip, 3))
}

// Replace returns err with replacement as its visible error while preserving
// both errors in the error tree.
func Replace(err, replacement error) error {
	if err == nil {
		return nil
	}
	if replacement == nil {
		return err
	}
	return &replacementError{replacement: replacement, err: err}
}

func (err *stackError) Error() string {
	return err.err.Error()
}

func (err *stackError) Unwrap() error {
	return err.err
}

func (err *stackError) Stack() string {
	err.once.Do(func() {
		err.stack = format(err.pcs)
		err.pcs = nil
	})
	return err.stack
}

func (err *replacementError) Error() string {
	return err.replacement.Error()
}

func (err *replacementError) Unwrap() []error {
	return []error{err.replacement, err.err}
}

func with(err error, skip int) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[Provider](err); ok {
		return err
	}
	return &stackError{err: err, pcs: capturePCs(skip, 4)}
}

func capturePCs(skip, base int) []uintptr {
	if skip < 0 {
		skip = 0
	}
	pcs := make([]uintptr, maxFrames)
	return pcs[:runtime.Callers(base+skip, pcs)]
}

func format(pcs []uintptr) string {
	if len(pcs) == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs)
	var stack strings.Builder
	for {
		frame, more := frames.Next()
		if stack.Len() > 0 {
			stack.WriteString("; ")
		}
		stack.WriteString(frame.Function)
		stack.WriteByte(' ')
		stack.WriteString(frame.File)
		stack.WriteByte(':')
		stack.WriteString(strconv.Itoa(frame.Line))
		if !more {
			return stack.String()
		}
	}
}
