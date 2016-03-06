package exception

import (
	"errors"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestCallerInfo(t *testing.T) {
	a := assert.New(t)

	stackTrace := callerInfo()
	a.NotEmpty(stackTrace)
}

func TestNew(t *testing.T) {
	a := assert.New(t)
	ex := AsException(New("this is a test"))
	a.Equal("this is a test", ex.Message())
	a.NotEmpty(ex.StackTrace())
	a.Nil(ex.InnerException())
}

func TestError(t *testing.T) {
	a := assert.New(t)

	ex := AsException(New("this is a test"))
	message := ex.Error()
	a.NotEmpty(message)
}

func TestNewf(t *testing.T) {
	a := assert.New(t)
	ex := AsException(Newf("%s", "this is a test"))
	a.Equal("this is a test", ex.Message())
	a.NotEmpty(ex.StackTrace())
	a.Nil(ex.InnerException())
}

func TestWrapError(t *testing.T) {
	a := assert.New(t)
	ex := AsException(WrapError(errors.New("this is a test")))
	a.Equal("this is a test", ex.Message())
	a.NotEmpty(ex.StackTrace())
	a.Nil(ex.InnerException())
}

func returnsNil() error {
	return nil
}

func returnsWrappedNil() error {
	return Wrap(nil)
}

func TestWrap(t *testing.T) {
	a := assert.New(t)

	err := errors.New("This is an error")
	ex := New("This is an exception")

	wrappedErr := Wrap(err)
	a.NotNil(wrappedErr)
	typedWrappedErr := AsException(wrappedErr)
	a.Equal("This is an error", typedWrappedErr.Message())

	wrappedEx := Wrap(ex)
	a.NotNil(wrappedEx)
	typedWrappedEx := AsException(wrappedEx)
	a.Equal("This is an exception", typedWrappedEx.Message())

	shouldBeNil := Wrap(nil)
	a.Nil(shouldBeNil)
	a.True(shouldBeNil == nil)

	var nilError error
	a.Nil(nilError)
	a.True(nilError == nil)
	shouldBeNil = Wrap(nilError)
	a.Nil(shouldBeNil)
	a.True(shouldBeNil == nil)

	shouldBeNil = Wrap(returnsNil())
	a.Nil(shouldBeNil)
	a.True(shouldBeNil == nil)

	shouldAlsoBeNil := returnsWrappedNil()
	a.Nil(shouldAlsoBeNil)
	a.True(shouldAlsoBeNil == nil)
}

func TestWrapMany(t *testing.T) {
	a := assert.New(t)

	err := errors.New("This is an error")
	ex1 := New("Exception1")
	ex2 := New("Exception2")

	combined := AsException(WrapMany(ex1, ex2, err))

	a.NotNil(combined)
	a.NotNil(combined.InnerException())
	a.NotNil(combined.InnerException().InnerException())
}
