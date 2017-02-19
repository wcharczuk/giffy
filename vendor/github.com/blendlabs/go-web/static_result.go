package web

import (
	"net/http"
	"path"
)

// NewStaticResultForSingleFile returns a static result for an individual file.
func NewStaticResultForSingleFile(filePath string) *StaticResult {
	file := path.Base(filePath)
	root := path.Dir(filePath)
	return &StaticResult{
		FilePath:   file,
		FileSystem: http.Dir(root),
	}
}

// NewStaticResultForDirectory returns a new static result for a directory and url path.
func NewStaticResultForDirectory(directoryPath, path string) *StaticResult {
	return &StaticResult{
		FilePath:   path,
		FileServer: http.FileServer(http.Dir(directoryPath)),
	}
}

// StaticResult represents a static output.
type StaticResult struct {
	FilePath   string
	FileSystem http.FileSystem
	FileServer http.Handler

	RewriteRules []*RewriteRule
	Headers      http.Header
}

// Render renders a static result.
func (sr StaticResult) Render(ctx *Ctx) error {
	filePath := sr.FilePath
	for _, rule := range sr.RewriteRules {
		if matched, newFilePath := rule.Apply(filePath); matched {
			filePath = newFilePath
		}
	}

	if sr.Headers != nil {
		for key, values := range sr.Headers {
			for _, value := range values {
				ctx.Response.Header().Add(key, value)
			}
		}
	}

	if sr.FileServer != nil {
		ctx.Request.URL.Path = filePath
		sr.FileServer.ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}

	return sr.serveStaticFile(ctx.Response, ctx.Request, sr.FileSystem, path.Clean(filePath))
}

func (sr StaticResult) serveStaticFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) error {
	f, err := fs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return err
	}

	http.ServeContent(w, r, name, d.ModTime(), f)
	return nil
}
