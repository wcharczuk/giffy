package web

import (
	"bytes"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestBufferedCompressedWriter(t *testing.T) {
	assert := assert.New(t)

	buf := bytes.NewBuffer(nil)
	mockedWriter := NewMockResponseWriter(buf)
	bufferedWriter := NewCompressedResponseWriter(mockedWriter)

	written, err := bufferedWriter.Write([]byte("ok"))
	assert.Nil(err)
	assert.NotZero(written)
}
