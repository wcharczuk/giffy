package web

import "net/http"

// Fileserver is a type that implements the basics of a fileserver.
type Fileserver interface {
	AddHeader(key, value string) error
	AddRewriteRule(route, match string, rewriteAction RewriteAction) error
	Headers() http.Header
	RewriteRules() []RewriteRule
	Action(*Ctx) Result
}
