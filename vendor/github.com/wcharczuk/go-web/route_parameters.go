package web

import "github.com/julienschmidt/httprouter"

// RouteParameters are parameters sourced from parsing the request path (route).
type RouteParameters map[string]string

// NewRouteParameters returns a new route parameters collection.
func NewRouteParameters(p httprouter.Params) RouteParameters {
	rp := RouteParameters{}
	for _, pv := range p {
		rp[pv.Key] = pv.Value
	}
	return rp
}

// Get gets a value for a key.
func (rp RouteParameters) Get(key string) string {
	return rp[key]
}

// Has returns if the collection has a key or not.
func (rp RouteParameters) Has(key string) bool {
	_, ok := rp[key]
	return ok
}

// Set stores a value for a key.
func (rp RouteParameters) Set(key, value string) {
	rp[key] = value
}
