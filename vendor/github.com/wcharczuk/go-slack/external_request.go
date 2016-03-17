package slack

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/blendlabs/go-request"
)

type mockedResponse struct {
	ResponseBody []byte
	StatusCode   int
	Error        error
}

var isMocked bool
var mocks map[string]mockedResponse

// Mock mocks a response for a given verb to a given url.
func Mock(verb string, url string, res mockedResponse) error {
	isMocked = true
	if mocks == nil {
		mocks = map[string]mockedResponse{}
	}
	storedURL := fmt.Sprintf("%s_%s", verb, url)
	mocks[storedURL] = res
	return nil
}

// MockResponseFromBytes mocks a request from a byte array response.
func MockResponseFromBytes(verb string, url string, statusCode int, response []byte) error {
	return Mock(verb, url, mockedResponse{ResponseBody: response, StatusCode: statusCode})
}

// MockResponseFromFile mocks a response from a file.
func MockResponseFromFile(verb string, url string, statusCode int, responseFilePath string) error {
	reader, readerErr := ioutil.ReadFile(responseFilePath)
	if readerErr != nil {
		return readerErr
	}
	return MockResponseFromBytes(verb, url, statusCode, reader)
}

// MockResponseFromString mocks a request from a string response.
func MockResponseFromString(verb string, url string, statusCode int, response string) error {
	return MockResponseFromBytes(verb, url, statusCode, []byte(response))
}

// ClearMockedResponses clears and disables response mocking
func ClearMockedResponses() {
	isMocked = false
	mocks = map[string]mockedResponse{}
}

// NewExternalRequest Creates a new external request
func NewExternalRequest() *request.HTTPRequest {
	req := request.NewHTTPRequest().WithMockedResponse(func(verb string, workingURL *url.URL) (bool, *request.HTTPResponseMeta, []byte, error) {
		if isMocked {
			storedURL := fmt.Sprintf("%s_%s", verb, workingURL.String())
			if mockResponse, ok := mocks[storedURL]; ok {
				meta := &request.HTTPResponseMeta{}
				meta.StatusCode = mockResponse.StatusCode
				meta.ContentLength = int64(len(mockResponse.ResponseBody))
				return true, meta, mockResponse.ResponseBody, mockResponse.Error
			}
			panic(fmt.Sprintf("attempted to make external request w/o mocking endpoint: %s %s", verb, workingURL.String()))
		} else {
			return false, nil, nil, nil
		}
	})

	return req
}
