package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
)

func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
	return nil
}

func WriteJson(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) (int, error) {
	bytes, err := json.Marshal(response)
	if err == nil {

		if requestConnectionHeader := r.Header.Get("Connection"); !util.IsEmpty(requestConnectionHeader) {
			if strings.ToLower(requestConnectionHeader) == "keep-alive" {
				w.Header().Set("Connection", "keep-alive")
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		count, write_error := w.Write(bytes)
		if count == 0 {
			return count, exception.New("`WriteJson` didn't write any bytes")
		}
		return count, exception.Wrap(write_error)
	} else {
		w.WriteHeader(500)
		w.Write([]byte{})
		return 0, exception.Wrap(err)
	}
}

func DeserializePostBody(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return exception.Wrap(err)
	}
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	return exception.Wrap(decoder.Decode(object))
}
