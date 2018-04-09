package logger

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	// Gigabyte is an SI unit.
	Gigabyte = 1 << 30
	// Megabyte is an SI unit.
	Megabyte = 1 << 20
	// Kilobyte is an SI unit.
	Kilobyte = 1 << 10
)

// Any is a helper alias to interface{}
type Any = interface{}

// GetIP gets the origin/client ip for a request.
// X-FORWARDED-FOR is checked. If multiple IPs are included the first one is returned
// X-REAL-IP is checked. If multiple IPs are included the first one is returned
// Finally r.RemoteAddr is used
// Only benevolent services will allow access to the real IP.
func GetIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-FOR", "X-REAL-IP"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// ColorizeByStatusCode returns a value colored by an http status code.
func ColorizeByStatusCode(statusCode int, value string) string {
	if statusCode >= http.StatusOK && statusCode < 300 { //the http 2xx range is ok
		return ColorGreen.Apply(value)
	} else if statusCode == http.StatusInternalServerError {
		return ColorRed.Apply(value)
	}
	return ColorYellow.Apply(value)
}

// ColorizeStatusCode colorizes a status code.
func ColorizeStatusCode(statusCode int) string {
	return ColorizeByStatusCode(statusCode, strconv.Itoa(statusCode))
}

// ParseFileSize returns a filesize.
func ParseFileSize(fileSizeValue string) (int64, error) {
	if len(fileSizeValue) == 0 {
		return 0, fmt.Errorf("empty filesize value")
	}

	units := strings.ToLower(fileSizeValue[len(fileSizeValue)-2:])
	value, err := strconv.ParseInt(fileSizeValue[:len(fileSizeValue)-2], 10, 64)
	if err != nil {
		return 0, err
	}
	switch units {
	case "gb":
		return value * Gigabyte, nil
	case "mb":
		return value * Megabyte, nil
	case "kb":
		return value * Kilobyte, nil
	}
	fullValue, err := strconv.ParseInt(fileSizeValue, 10, 64)
	if err != nil {
		return 0, err
	}
	return fullValue, nil
}

// FormatFileSize returns a string representation of a file size in bytes.
func FormatFileSize(sizeBytes int64) string {
	if sizeBytes >= 1<<30 {
		return strconv.FormatInt(sizeBytes/Gigabyte, 10) + "gb"
	} else if sizeBytes >= 1<<20 {
		return strconv.FormatInt(sizeBytes/Megabyte, 10) + "mb"
	} else if sizeBytes >= 1<<10 {
		return strconv.FormatInt(sizeBytes/Kilobyte, 10) + "kb"
	}
	return strconv.FormatInt(sizeBytes, 10)
}
