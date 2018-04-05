package web

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestViewResultRender(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc := NewCtx(NewMockResponseWriter(buffer), nil, nil, nil)

	testView := template.New("testView")
	testView.Parse("{{.ViewModel.Text}}")

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   testView,
	}

	err := vr.Render(rc)
	assert.Nil(err)

	assert.NotZero(buffer.Len())
	assert.True(strings.Contains(buffer.String(), "bar"))
}

func TestViewResultRenderError(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc := NewCtx(NewMockResponseWriter(buffer), nil, nil, nil)

	testView := template.New("testView")
	testView.Parse("{{.ViewModel.Foo}}")

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   testView,
	}

	err := vr.Render(rc)
	assert.NotNil(err)
	assert.Zero(buffer.Len())
}

func TestViewResultRenderErrorTemplate(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc := NewCtx(NewMockResponseWriter(buffer), nil, nil, nil)

	views := template.New("main")

	errorTemplate := views.New(DefaultTemplateNameInternalError)
	_, err := errorTemplate.Parse("{{.}}")
	assert.Nil(err)

	testView := views.New("testView")
	_, err = testView.Parse("Foo: {{.ViewModel.Foo}}")
	assert.Nil(err)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   testView,
	}

	err = vr.Render(rc)
	assert.NotNil(err)
}

func TestViewResultErrorNestedViews(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc := NewCtx(NewMockResponseWriter(buffer), nil, nil, nil)

	views := template.New("main")

	errorTemplate := views.New(DefaultTemplateNameInternalError)
	_, err := errorTemplate.Parse("{{.}}")
	assert.Nil(err)

	outerView := views.New("outerView")
	outerView.Parse("Inner: {{ template \"innerView\" . }}")

	testView := views.New("innerView")
	_, err = testView.Parse("Foo: {{.ViewModel.Foo}}")
	assert.Nil(err)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   outerView,
	}

	err = vr.Render(rc)
	assert.NotNil(err)
}
