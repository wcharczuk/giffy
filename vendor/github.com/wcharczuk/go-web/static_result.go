package web

import "net/http"

// StaticResult represents a static output.
type StaticResult struct {
	FilePath   string
	FileServer http.Handler

	RewriteRules []*RewriteRule
	Headers      http.Header
}

// Render renders a static result.
func (sr StaticResult) Render(w http.ResponseWriter, r *http.Request) error {
	filePath := sr.FilePath
	for _, rule := range sr.RewriteRules {
		if matched, newFilePath := rule.Apply(filePath); matched {
			filePath = newFilePath
		}
	}

	if sr.Headers != nil {
		for key, values := range sr.Headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}

	r.URL.Path = filePath
	sr.FileServer.ServeHTTP(w, r)
	return nil
}
