package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core"
)

const (
	// RequestLogItemPrefixClient is the prefix for client items.
	RequestLogItemPrefixClient = "c"
	// RequestLogItemPrefixServer is the prefix for server items.
	RequestLogItemPrefixServer = "s"
	// RequestLogItemPrefixRemote is the prefix for remote items.
	RequestLogItemPrefixRemote = "r"
	// RequestLogItemPrefixClientServer is the prefix for combination client and server items.
	RequestLogItemPrefixClientServer = "cs"
	// RequestLogItemPrefixServerClient is the prefix for combination client and server items.
	RequestLogItemPrefixServerClient = "sc"

	// RequestLogItemDateTime is the timestamp item.
	RequestLogItemDateTime = "datetime" //w3c has separate date and time fields, we just use one.
	// RequestLogItemTimeTaken is the elapsed time item.
	RequestLogItemTimeTaken = "time-taken"
	// RequestLogItemBytes is the size of the resulting message.
	RequestLogItemBytes = "bytes"
	// RequestLogItemCached is a flag indicating if the response was cached.
	RequestLogItemCached = "cached"

	// RequestLogItemIP requires a prefix.
	RequestLogItemIP = "ip"
	// RequestLogItemDNS requires a prefix.
	RequestLogItemDNS = "dns"
	// RequestLogItemStatus requires a prefix.
	RequestLogItemStatus = "status" //status code ... why does this need a prefix.
	// RequestLogItemComment requires a prefix.
	RequestLogItemComment = "comment"
	// RequestLogItemMethod requires a prefix.
	RequestLogItemMethod = "method"
	// RequestLogItemURI requires a prefix.
	RequestLogItemURI = "uri"
	// RequestLogItemURIStem requires a prefix.
	RequestLogItemURIStem = "uri-stem"
	// RequestLogItemURIQuery requires a prefix.
	RequestLogItemURIQuery = "uri-query"
)

// RequestLogPrefixes are prefixes for log item.
var RequestLogPrefixes = []string{
	RequestLogItemPrefixClientServer,
	RequestLogItemPrefixServerClient,
	RequestLogItemPrefixClient,
	RequestLogItemPrefixServer,
	RequestLogItemPrefixRemote,
}

// RequestLogItemsWithPrefix are log items that require a prefix.
var RequestLogItemsWithPrefix = []string{
	RequestLogItemIP,
	RequestLogItemDNS,
	RequestLogItemStatus,
	RequestLogItemComment,
	RequestLogItemMethod,
	RequestLogItemURI,
	RequestLogItemURIStem,
	RequestLogItemURIQuery,
}

func getLoggingTimestamp() string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return util.Color(timestamp, util.ColorGray)
}

func formatFileSize(sizeBytes int) string {
	if sizeBytes >= 1<<30 {
		return fmt.Sprintf("%dgB", sizeBytes/(1<<30))
	} else if sizeBytes >= 1<<20 {
		return fmt.Sprintf("%dmB", sizeBytes/(1<<20))
	} else if sizeBytes >= 1<<10 {
		return fmt.Sprintf("%dkB", sizeBytes/(1<<10))
	}
	return fmt.Sprintf("%dB", sizeBytes)
}

func escapeRequestLogOutput(format string, context *HTTPContext) string {
	output := format

	//log item: datetime
	dateTime := getLoggingTimestamp()
	output = strings.Replace(output, RequestLogItemDateTime, dateTime, -1)

	//log item: time-taken
	timeTakenStr := fmt.Sprintf("%v", context.elapsed())
	output = strings.Replace(output, RequestLogItemTimeTaken, timeTakenStr, -1)

	//log item: bytes
	contentLengthStr := fmt.Sprintf("%v", formatFileSize(context.getContentLength()))
	output = strings.Replace(output, RequestLogItemBytes, contentLengthStr, -1)

	//log item: cached
	output = strings.Replace(output, RequestLogItemCached, "false", -1)

	clientIP := util.GetIP(context.Request)
	output = strings.Replace(output, RequestLogItemPrefixClient+"-"+RequestLogItemIP, clientIP, -1)

	serverIP := core.ConfigLocalIP()
	output = strings.Replace(output, RequestLogItemPrefixServer+"-"+RequestLogItemIP, serverIP, -1)

	status := util.Color(util.IntToString(context.getStatusCode()), util.ColorYellow)
	if context.getStatusCode() == http.StatusOK {
		status = util.Color(util.IntToString(context.getStatusCode()), util.ColorGreen)
	} else if context.getStatusCode() == http.StatusInternalServerError {
		status = util.Color(util.IntToString(context.getStatusCode()), util.ColorRed)
	}

	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemStatus, status, -1)
	}
	output = strings.Replace(output, RequestLogItemStatus, status, -1)

	method := util.Color(strings.ToUpper(context.Request.Method), util.ColorBlue)
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemMethod, method, -1)
	}
	output = strings.Replace(output, RequestLogItemMethod, method, -1)

	fullURI := context.Request.URL.String()
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemURI, fullURI, -1)
	}
	output = strings.Replace(output, RequestLogItemURI, fullURI, -1)

	uriPath := context.Request.URL.Path
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemURIStem, uriPath, -1)
	}
	output = strings.Replace(output, RequestLogItemURIStem, uriPath, -1)

	rawQuery := context.Request.URL.RawQuery
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemURIQuery, rawQuery, -1)
	}
	output = strings.Replace(output, RequestLogItemURIQuery, rawQuery, -1)

	return output
}

// LogError logs an error.
func LogError(err error) {
	Logf("%v", err)
}

// Log writes to the log output.
func Log(a ...interface{}) {
	timestamp := getLoggingTimestamp()
	output := fmt.Sprint(a...)
	fmt.Printf("%s %s\n", timestamp, output)
}

// Logf writes to the log with a given format.
func Logf(format string, a ...interface{}) {
	timestamp := getLoggingTimestamp()
	output := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s\n", timestamp, output)
}
