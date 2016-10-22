package web

import "net/http"

// RawResult is for when you just want to dump bytes.
type RawResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Render renders the result.
func (rr *RawResult) Render(w http.ResponseWriter, r *http.Request) error {
	if len(rr.ContentType) != 0 {
		w.Header().Set("Content-Type", rr.ContentType)
	}
	if rr.StatusCode == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(rr.StatusCode)
	}
	_, err := w.Write(rr.Body)
	return err
}
