package web

import (
	"io/ioutil"
	"net/http"
	"os"
)

// StaticResult represents a static output.
type StaticResult struct {
	FilePath string
}

// Render renders a static result.
func (sr StaticResult) Render(ctx *RequestContext) error {
	f, err := os.Open(sr.FilePath)
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	ctx.Response.Header().Set("Content-Type", http.DetectContentType(contents))
	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.Write(contents)
	return nil
}
