package logging

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blendlabs/connectivity/core"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core/web"
)

const (
	//the below are for W3C extended log format
	RequestLogItemPrefixClient       = "c"
	RequestLogItemPrefixServer       = "s"
	RequestLogItemPrefixRemote       = "r"
	RequestLogItemPrefixClientServer = "cs"
	RequestLogItemPrefixServerClient = "sc"

	//these can just appear as a token
	RequestLogItemDateTime  = "datetime"   //w3c has separate date and time fields, we just use one.
	RequestLogItemTimeTaken = "time-taken" //we call this "with timing"
	RequestLogItemBytes     = "bytes"      //response content-length
	RequestLogItemCached    = "cached"     //will always return false

	//these require a prefix
	RequestLogItemIP       = "ip"
	RequestLogItemDNS      = "dns"
	RequestLogItemStatus   = "status" //status code ... why does this need a prefix.
	RequestLogItemComment  = "comment"
	RequestLogItemMethod   = "method"
	RequestLogItemURI      = "uri"
	RequestLogItemURIStem  = "uri-stem"
	RequestLogItemURIQuery = "uri-query"
)

var RequestLogPrefixes = []string{
	RequestLogItemPrefixClientServer,
	RequestLogItemPrefixServerClient,
	RequestLogItemPrefixClient,
	RequestLogItemPrefixServer,
	RequestLogItemPrefixRemote,
}

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
	timestamp_str := time.Now().UTC().Format(time.RFC3339)
	return util.Color(timestamp_str, util.COLOR_GRAY)
}

func escapeRequestLogOutput(format string, context *web.APIContext) string {
	output := format

	//log item: datetime
	dateTime := getLoggingTimestamp()
	output = strings.Replace(output, RequestLogItemDateTime, dateTime, -1)

	//log item: time-taken
	timeTakenStr := fmt.Sprintf("%v", context.Elapsed())
	output = strings.Replace(output, RequestLogItemTimeTaken, timeTakenStr, -1)

	//log item: bytes
	contentLengthStr := fmt.Sprintf("%v", formatFileSize(context.ContentLength()))
	output = strings.Replace(output, RequestLogItemBytes, contentLengthStr, -1)

	//log item: cached
	output = strings.Replace(output, RequestLogItemCached, "false", -1)

	clientIp := util.GetIP(context.Request)
	output = strings.Replace(output, RequestLogItemPrefixClient+"-"+RequestLogItemIP, clientIp, -1)

	serverIp := core.ConfigLocalIP()
	output = strings.Replace(output, RequestLogItemPrefixServer+"-"+RequestLogItemIP, serverIp, -1)

	status := util.Color(util.IntToString(context.StatusCode()), util.COLOR_YELLOW)
	if context.StatusCode() == http.StatusOK {
		status = util.Color(util.IntToString(context.StatusCode()), util.COLOR_GREEN)
	} else if context.StatusCode() == http.StatusInternalServerError {
		status = util.Color(util.IntToString(context.StatusCode()), util.COLOR_RED)
	}

	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemStatus, status, -1)
	}
	output = strings.Replace(output, RequestLogItemStatus, status, -1)

	method := util.Color(strings.ToUpper(context.Request.Method), util.COLOR_BLUE)
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemMethod, method, -1)
	}
	output = strings.Replace(output, RequestLogItemMethod, method, -1)

	fullUri := context.Request.URL.String()
	for _, prefix := range RequestLogPrefixes {
		output = strings.Replace(output, prefix+"-"+RequestLogItemURI, fullUri, -1)
	}
	output = strings.Replace(output, RequestLogItemURI, fullUri, -1)

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

func LogRequest(format string, context *web.APIContext) {
	fmt.Println(escapeRequestLogOutput(format, context))
}

func Log(a ...interface{}) {
	timestamp := getLoggingTimestamp()
	output := fmt.Sprint(a...)
	fmt.Printf("%s %s\n", timestamp, output)
}

func Logf(format string, a ...interface{}) {
	timestamp := getLoggingTimestamp()
	output := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s\n", timestamp, output)
}
