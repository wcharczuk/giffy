package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-logger"
)

func agent() *logger.Logger {
	return logger.None()
}

func TestViewResultProviderNotFound(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(nil, NewViewCache()).NotFound()
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusNotFound, typed.StatusCode)
}

func TestViewResultProviderNotAuthorized(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(nil, NewViewCache()).NotAuthorized()
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusForbidden, typed.StatusCode)
}

func TestViewResultProviderInternalError(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(nil, NewViewCache()).InternalError(exception.New("Test"))
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)
}

func TestViewResultProviderInternalErrorWritesToLogger(t *testing.T) {
	assert := assert.New(t)

	logBuffer := bytes.NewBuffer([]byte{})
	log := logger.New(logger.Fatal).WithWriter(logger.NewTextWriter(logBuffer))
	defer log.Close()

	assert.True(log.IsEnabled(logger.Fatal))

	wg := sync.WaitGroup{}
	wg.Add(1)

	app := New().WithLogger(log)
	app.Logger().Listen(logger.Fatal, "foo", func(e logger.Event) {
		defer wg.Done()
	})

	result := NewViewResultProvider(log, NewViewCache()).InternalError(exception.New("Test"))
	assert.NotNil(result)

	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)

	wg.Wait()
	log.Drain()
	assert.NotZero(logBuffer.Len())
}

func TestViewResultProviderBadRequest(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(nil, NewViewCache()).BadRequest(fmt.Errorf("test"))
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusBadRequest, typed.StatusCode)
}

type testViewModel struct {
	Text string
}

func TestViewResultProviderView(t *testing.T) {
	assert := assert.New(t)

	testView := template.New("testView")
	testView.Parse("{{.Text}}")

	provider := NewViewResultProvider(nil, NewViewCache())
	provider.views.SetTemplates(testView)
	result := provider.View("testView", testViewModel{Text: "foo"})

	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusOK, typed.StatusCode)
}
