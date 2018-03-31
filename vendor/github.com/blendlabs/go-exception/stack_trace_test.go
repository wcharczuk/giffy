package exception

import (
	"fmt"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestGetStackTrace(t *testing.T) {
	assert := assert.New(t)

	assert.NotEmpty(GetStackTrace())
}

func TestStackStrings(t *testing.T) {
	assert := assert.New(t)

	stack := []string{
		"foo",
		"bar",
		"baz",
	}

	stackStrings := StackStrings(stack)

	assert.Equal("\nfoo\nbar\nbaz", fmt.Sprintf("%+v", stackStrings))
	assert.Equal("[]string{\"foo\", \"bar\", \"baz\"}", fmt.Sprintf("%#v", stackStrings))
	assert.Equal("\nfoo\nbar\nbaz", fmt.Sprintf("%v", stackStrings))
	assert.Equal("\nfoo\nbar\nbaz", fmt.Sprintf("%s", stackStrings))
}

func TestExceptionWithStackStrings(t *testing.T) {
	assert := assert.New(t)

	stack := []string{
		"foo",
		"bar",
		"baz",
	}

	ex := New("foo").WithStack(StackStrings(stack))

	values := ex.Decompose()
	assert.NotEmpty(values["Stack"])

	assert.NotNil(ex.Stack())
}
