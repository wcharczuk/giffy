package util

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/blendlabs/go-exception"
)

// ReadFile reads a file as a string.
func ReadFile(path string) string {
	bytes, _ := ioutil.ReadFile(path)
	return string(bytes)
}

// DeserializeJSON unmarshals an object from JSON.
func DeserializeJSON(object interface{}, body string) error {
	decoder := json.NewDecoder(bytes.NewBufferString(body))
	return exception.Wrap(decoder.Decode(object))
}

// DeserializeJSONFromReader unmashals an object from a json Reader.
func DeserializeJSONFromReader(object interface{}, body io.Reader) error {
	decoder := json.NewDecoder(body)
	return exception.Wrap(decoder.Decode(object))
}

// DeserializeJSONFromReadCloser unmashals an object from a json ReadCloser.
func DeserializeJSONFromReadCloser(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return exception.Wrap(err)
	}

	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	return exception.Wrap(decoder.Decode(object))
}

// SerializeJSON marshals an object to json.
func SerializeJSON(object interface{}) string {
	b, _ := json.Marshal(object)
	return string(b)
}

// SerializeJSONPretty marshals an object to json with formatting whitespace.
func SerializeJSONPretty(object interface{}, prefix, indent string) string {
	b, _ := json.MarshalIndent(object, prefix, indent)
	return string(b)
}

// SerializeJSONAsReader marshals an object to json as a reader.
func SerializeJSONAsReader(object interface{}) io.Reader {
	b, _ := json.Marshal(object)
	return bytes.NewBufferString(string(b))
}
