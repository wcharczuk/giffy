package web

import "regexp"

// RewriteRule is a rule for re-writing incoming static urls.
type RewriteRule struct {
	MatchExpression string
	expr            *regexp.Regexp
	Action          func(matchedPieces ...string) string
}

// Apply runs the filter, returning a bool if it matched, and the resulting path.
func (rr RewriteRule) Apply(filePath string) (bool, string) {
	if rr.expr.MatchString(filePath) {
		pieces := extractSubMatches(rr.expr, filePath)
		return true, rr.Action(pieces...)
	}

	return false, filePath
}

// ExtractSubMatches returns sub matches for an expr because go's regexp library is weird.
func extractSubMatches(re *regexp.Regexp, corpus string) []string {
	allResults := re.FindAllStringSubmatch(corpus, -1)
	results := []string{}
	for _, resultSet := range allResults {
		for _, result := range resultSet {
			results = append(results, result)
		}
	}

	return results
}