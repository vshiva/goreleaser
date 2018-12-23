// Package errors provides a central interface to handling
// all errors in the Athens domain. It closely follows Upspin's error
// handling with a few differences that reflect the system design
// of the Athen's architecture. If you're unfamiliar with Upspin's error
// handling, we recommend that you read this article first before
// coming back here https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html.
// Athen's errors are central around dealing with modules. So every error
// will most likely carry a Module and a Version inside of them. Furthermore,
// because Athens is designed to run on multiple clouds, we have to design
// our errors and our logger to be friendly with each other. Therefore,
// the logger's SystemError method, although it accepts any type of error,
// it knows how to deal with errors constructed from this package in a debuggable way.
// To construct an Athens error, call the errors.E function. The E function takes
// an Op and a variadic interface{}, but the values of the Error struct are what you can
// pass to it. Values such as the error Kind, Module, Version, Error Message,
// and Seveirty (seriousness of an error) are all optional. The only truly required value is
// the errors.Op so you can construct a traceable stack that leads to where
// the error happened. However, adding more information can help catch an issue
// quicker and would help Cloud Log Monitoring services be more efficient to maintainers
// as you can run queries on Athens Errors such as "Give me all errors of KindUnexpected"
// or "Give me all errors where caused by a particular Module"
package errors

import (
	"errors"
	"fmt"
	"runtime"
)

const (
	// KindPipeSkipped happens when a pipe is skipped.
	KindPipeSkipped int = iota + 1
	// KindBuildError is a build error
	KindBuildError
	// KindSemVerError is an error related to semantic versioning
	KindSemVerError
	// KindUnexpected is an unexpected error
	KindUnexpected
)

// Error is a GoReleaser system error.
// It carries information and behavior
// as to what caused this error so that
// callers can implement logic around it.
type Error struct {
	Kind int
	Op   Op
	Err  error
}

// Error returns the underlying error's
// string message. The logger takes care
// of filling out the stack levels and
// extra information.
func (e Error) Error() string {
	return e.Err.Error()
}

// Op describes any independent function or
// method in GoReleaser. A series of operations
// forms a more readable stack trace.
type Op string

func (o Op) String() string {
	return string(o)
}

func Skip(op Op, args ...interface{}) error {
	args = append(args, KindPipeSkipped)
	return E(op, args...)
}

func IsSkip(err error) bool {
	return Kind(err) == KindPipeSkipped
}

// E is a helper function to construct an Error type
// Operation always comes first, module path and version
// come second, they are optional. Args must have at least
// an error or a string to describe what exactly went wrong.
// You can optionally pass a Logrus severity to indicate
// the log level of an error based on the context it was constructed in.
func E(op Op, args ...interface{}) error {
	e := Error{Op: op}
	if len(args) == 0 {
		msg := "errors.E called with 0 args"
		_, file, line, ok := runtime.Caller(1)
		if ok {
			msg = fmt.Sprintf("%v - %v:%v", msg, file, line)
		}
		e.Err = errors.New(msg)
	}
	for _, a := range args {
		switch a := a.(type) {
		case error:
			e.Err = a
		case string:
			if e.Err != nil {
				e.Err = fmt.Errorf("%s: %v", a, e.Err)
			} else {
				e.Err = errors.New(a)
			}
		case int:
			e.Kind = a
		}
	}
	// if no err is passed, assume pkg/errors.Wrap behavior of returning nil
	if e.Err == nil {
		return nil
	}
	return e
}

// Kind recursively searches for the
// first error kind it finds.
func Kind(err error) int {
	e, ok := err.(Error)
	if !ok {
		return KindUnexpected
	}

	if e.Kind != 0 {
		return e.Kind
	}

	return Kind(e.Err)
}

// KindText returns a friendly string
// of the Kind type. Since we use http
// status codes to represent error kinds,
// this method just deferrs to the net/http
// text representations of statuses.
func KindText(err error) string {
	switch Kind(err) {
	case KindPipeSkipped:
		return "pipe was skipped"
	case KindBuildError:
		return "build failed"
	case KindSemVerError:
		return "invalid semantic version"
	}
	return "unexpected error"
}

// Ops aggregates the error's operation
// with all the embedded errors' operations.
// This way you can construct a queryable
// stack trace.
func Ops(err error) []Op {
	gerr, ok := err.(Error)
	if !ok {
		return []Op{"undefined"}
	}

	ops := []Op{gerr.Op}
	for {
		embeddedErr, ok := gerr.Err.(Error)
		if !ok {
			break
		}

		ops = append(ops, embeddedErr.Op)
		gerr = embeddedErr
	}

	return ops
}
