package web

import "github.com/blend/go-sdk/util"

// ViewCacheConfig is a config for the view cache.
type ViewCacheConfig struct {
	Cached *bool    `json:"cached" yaml:"cached" env:"VIEW_CACHE_ENABLED"`
	Paths  []string `json:"paths" yaml:"paths" env:"VIEW_CACHE_PATHS,csv"`
}

// GetCached returns if the viewcache should store templates in memory or read from disk.
func (vcc ViewCacheConfig) GetCached(defaults ...bool) bool {
	return util.Coalesce.Bool(vcc.Cached, true, defaults...)
}

// GetPaths returns default view paths.
func (vcc ViewCacheConfig) GetPaths(defaults ...[]string) []string {
	return util.Coalesce.Strings(vcc.Paths, nil, defaults...)
}
