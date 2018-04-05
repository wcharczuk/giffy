package exception

import (
	"encoding/json"
	"fmt"
	"io"
)

// New returns a new exception with a call stack.
func New(classArgs ...interface{}) Exception {
	return &Ex{
		class: fmt.Sprint(classArgs...),
		stack: callers(),
	}
}

// Newf returns a new exception by `Sprintf`ing the format and the args.
func Newf(classFormat string, args ...interface{}) Exception {
	return &Ex{
		class: fmt.Sprintf(classFormat, args...),
		stack: callers(),
	}
}

// NewFromErr returns a new exception from an error.
func NewFromErr(err error) Exception {
	return &Ex{
		inner: err,
		class: err.Error(),
		stack: callers(),
	}
}

// Nest nests an arbitrary number of exceptions.
func Nest(err ...error) Exception {
	var ex *Ex
	var last *Ex
	var didSet bool

	for _, e := range err {
		if e != nil {
			var wrappedEx *Ex
			if typedEx, isTyped := e.(*Ex); !isTyped {
				wrappedEx = &Ex{
					class: e.Error(),
					stack: callers(),
				}
			} else {
				wrappedEx = typedEx
			}

			if wrappedEx != ex {
				if ex == nil {
					ex = wrappedEx
					last = wrappedEx
				} else {
					last.inner = wrappedEx
					last = wrappedEx
				}
				didSet = true
			}
		}
	}
	if didSet {
		return ex
	}
	return nil
}

// Wrap wraps an exception, will return error-typed `nil` if the exception is nil.
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	if typed, isTyped := err.(*Ex); isTyped {
		return typed
	}
	return NewFromErr(err)
}

// Exception is an exception.
type Exception interface {
	error
	fmt.Formatter
	json.Marshaler

	WithClass(string) Exception
	Class() string
	WithMessagef(string, ...interface{}) Exception
	Message() string
	WithInner(error) Exception
	Inner() error
	WithStack(StackTrace) Exception
	Stack() StackTrace

	Decompose() map[string]interface{}
}

// Ex is an error with a stack trace.
type Ex struct {
	// Class disambiguates between errors, it can be used to identify the type of the error.
	class string
	// Message adds further detail to the error, and shouldn't be used for disambiguation.
	message string
	// Inner holds the original error in cases where we're wrapping an error with a stack trace.
	inner error
	// Stack is the call stack frames used to create the stack output.
	stack StackTrace
}

// Format allows for conditional expansion in printf statements
// based on the token and flags used.
// 	%+v : class + message + stack
// 	%v, %c : class
// 	%m : message
// 	%t : stack
func (e *Ex) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if len(e.class) > 0 {
				fmt.Fprintf(s, "%s", e.class)
				if len(e.message) > 0 {
					fmt.Fprintf(s, "\nmessage: %s", e.message)
				}
			} else if len(e.message) > 0 {
				io.WriteString(s, e.message)
			}

			e.stack.Format(s, verb)
			return
		} else if s.Flag('-') {
			e.stack.Format(s, verb)
			return
		}

		if len(e.class) > 0 {
			io.WriteString(s, e.class)
			if len(e.message) > 0 {
				fmt.Fprintf(s, "\nmessage: %s", e.message)
			}
		} else if len(e.message) > 0 {
			io.WriteString(s, e.message)
		}
		return
	case 'c':
		io.WriteString(s, e.class)
	case 'm':
		io.WriteString(s, e.message)
	case 'q':
		fmt.Fprintf(s, "%q", e.message)
	}
}

// Error implements the `error` interface
func (e *Ex) Error() string { return e.class }

// Decompose breaks the exception down to be marshalled into an intermediate format.
func (e *Ex) Decompose() map[string]interface{} {
	values := map[string]interface{}{}
	values["Class"] = e.class
	values["Message"] = e.message
	if e.stack != nil {
		values["Stack"] = e.Stack().Strings()
	}
	if e.inner != nil {
		if typed, isTyped := e.inner.(*Ex); isTyped {
			values["Inner"] = typed.Decompose()
		} else {
			values["Inner"] = e.inner.Error()
		}
	}
	return values
}

// MarshalJSON is a custom json marshaler.
func (e *Ex) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Decompose())
}

// WithClass sets the exception class and returns the exepction.
func (e *Ex) WithClass(class string) Exception {
	e.class = class
	return e
}

// Class returns the exception class.
func (e *Ex) Class() string {
	return e.class
}

// WithInner sets the inner and returns the exception.
func (e *Ex) WithInner(err error) Exception {
	e.inner = err
	return e
}

// Inner returns the nested exception.
func (e *Ex) Inner() error {
	return e.inner
}

// Message returns just the message, it is effectively
// an alias to .Error()
func (e *Ex) Message() string { return e.message }

// WithMessagef sets the message based on a format and args, and returns the exception.
func (e *Ex) WithMessagef(format string, args ...interface{}) Exception {
	e.message = fmt.Sprintf(format, args...)
	return e
}

// WithStack sets the stack.
func (e *Ex) WithStack(stack StackTrace) Exception {
	e.stack = stack
	return e
}

// Stack returns the stack provider.
// This is typically the runtime []uintptr or []string if restored after the fact.
func (e *Ex) Stack() StackTrace {
	return e.stack
}

// StackString returns the stack trace as a string.
func (e *Ex) StackString() string {
	return fmt.Sprintf("%v", e.stack)
}
