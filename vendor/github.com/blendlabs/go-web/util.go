package web

import (
	"crypto/hmac"
	cryptoRand "crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/blendlabs/go-exception"
)

// NestMiddleware reads the middleware variadic args and organizes the calls recursively in the order they appear.
func NestMiddleware(action Action, middleware ...Middleware) Action {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b Middleware) Middleware {
		if b == nil {
			return a
		}
		return func(action Action) Action {
			return a(b(action))
		}
	}

	var metaAction Middleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// WriteNoContent writes http.StatusNoContent for a request.
func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
	return nil
}

// WriteRawContent writes raw content for the request.
func WriteRawContent(w http.ResponseWriter, statusCode int, content []byte) error {
	w.WriteHeader(statusCode)
	_, err := w.Write(content)
	return exception.Wrap(err)
}

// WriteJSON marshalls an object to json.
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	return exception.Wrap(json.NewEncoder(w).Encode(response))
}

// WriteXML marshalls an object to json.
func WriteXML(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeXML)
	w.WriteHeader(statusCode)
	return exception.Wrap(xml.NewEncoder(w).Encode(response))
}

// DeserializeReaderAsJSON deserializes a post body as json to a given object.
func DeserializeReaderAsJSON(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	return exception.Wrap(json.NewDecoder(body).Decode(object))
}

// LocalIP returns the local server ip.
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	return String.SecureRandom(LenSessionID)
}

// SignSessionID returns a new secure session id.
func SignSessionID(sessionID string, key []byte) ([]byte, error) {
	mac := hmac.New(sha512.New, key)
	_, err := mac.Write([]byte(sessionID))
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}

// EncodeSignSessionID returns a new secure session id base64 encoded..
func EncodeSignSessionID(sessionID string, key []byte) (string, error) {
	signed, err := SignSessionID(sessionID, key)
	if err != nil {
		return "", err
	}
	return Base64.Encode(signed), nil
}

// GenerateCryptoKey generates a cryptographic key.
func GenerateCryptoKey(keySize int) []byte {
	key := make([]byte, keySize)
	io.ReadFull(cryptoRand.Reader, key)
	return key
}

// GenerateSHA512Key generates a crypto key for SHA512 hashing.
func GenerateSHA512Key() []byte {
	return GenerateCryptoKey(64)
}

// PortFromBindAddr returns a port number as an integer from a bind addr.
func PortFromBindAddr(bindAddr string) int32 {
	if len(bindAddr) == 0 {
		return 0
	}
	parts := strings.SplitN(bindAddr, ":", 2)
	if len(parts) == 0 {
		return 0
	}
	if len(parts) < 2 {
		return ParseInt32(parts[0])
	}
	return ParseInt32(parts[1])
}

// ParseInt32 parses an int32.
func ParseInt32(v string) int32 {
	parsed, _ := strconv.Atoi(v)
	return int32(parsed)
}
