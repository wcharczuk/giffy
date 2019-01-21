package request

// ResponseHandler is a receiver for `OnResponse`.
type ResponseHandler func(req *Request, res *ResponseMeta, content []byte)

// Handler is a receiver for `OnRequest`.
type Handler func(req *Request)

// MockedResponseProvider is a mocked response provider.
type MockedResponseProvider func(*Request) *MockedResponse

// Deserializer is a function that does things with the response body.
type Deserializer func(body []byte) error

// Serializer is a function that turns an object into raw data.
type Serializer func(value interface{}) ([]byte, error)
