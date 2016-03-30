package web

// RouteParameters are parameters sourced from parsing the request path (route).
type RouteParameters map[string]string

func NewRouteParameters() RouteParameters {
	return RouteParameters{}
}
