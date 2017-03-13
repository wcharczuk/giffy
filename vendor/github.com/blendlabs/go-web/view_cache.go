package web

import (
	"fmt"
	"html/template"
	"time"
)

// NewViewCache returns a new view cache.
func NewViewCache() *ViewCache {
	return &ViewCache{
		viewFuncMap: viewUtils(),
		viewCache:   template.New(""),
		enabled:     true,
	}
}

// NewViewCacheWithTemplates creates a new view cache wrapping the templates.
func NewViewCacheWithTemplates(templates *template.Template) *ViewCache {
	return &ViewCache{
		viewFuncMap: viewUtils(),
		viewCache:   templates,
		enabled:     true,
	}
}

// ViewCache is the cached views used in view results.
type ViewCache struct {
	viewFuncMap template.FuncMap
	viewPaths   []string
	viewCache   *template.Template
	enabled     bool
}

// SetEnabled sets the view cache to either cache views or re-parse every request.
func (vc *ViewCache) SetEnabled(enabled bool) {
	vc.enabled = enabled
}

// Enabled indicates if the cache is enabled, or if it re-parses views every request.
func (vc *ViewCache) Enabled() bool {
	return vc.enabled
}

// Initialize caches templates by path.
func (vc *ViewCache) Initialize() error {
	if len(vc.viewPaths) == 0 {
		return nil
	}

	views, err := vc.Parse()
	if err != nil {
		return err
	}
	vc.viewCache = views
	return nil
}

// Parse parses the view tree.
func (vc *ViewCache) Parse() (*template.Template, error) {
	return template.New("").Funcs(vc.viewFuncMap).ParseFiles(vc.viewPaths...)
}

// AddPaths adds paths to the view collection.
func (vc *ViewCache) AddPaths(paths ...string) {
	vc.viewPaths = append(vc.viewPaths, paths...)
}

// SetPaths sets the view paths outright.
func (vc *ViewCache) SetPaths(paths ...string) {
	vc.viewPaths = paths
}

// Paths returns the view paths.
func (vc *ViewCache) Paths() []string {
	return vc.viewPaths
}

// FuncMap returns the global view func map.
func (vc *ViewCache) FuncMap() template.FuncMap {
	return vc.viewFuncMap
}

// Templates gets the view cache for the app.
func (vc *ViewCache) Templates() *template.Template {
	return vc.viewCache
}

// SetTemplates sets the view cache for the app.
func (vc *ViewCache) SetTemplates(viewCache *template.Template) {
	vc.viewCache = viewCache
}

func viewUtils() template.FuncMap {
	return template.FuncMap{
		"short": func(t time.Time) string {
			return t.Format("1/02/2006 3:04:05 PM")
		},
		"medium": func(t time.Time) string {
			return t.Format("Jan 02, 2006 3:04:05 PM")
		},
		"kitchen": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
		"money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},
	}
}
