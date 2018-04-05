package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
)

// NewViewCache returns a new view cache.
func NewViewCache() *ViewCache {
	return &ViewCache{
		viewFuncMap: viewUtils(),
		viewCache:   template.New(""),
		cached:      true,
	}
}

// NewViewCacheFromConfig returns a new view cache from a config.
func NewViewCacheFromConfig(cfg *ViewCacheConfig) *ViewCache {
	return &ViewCache{
		viewFuncMap: viewUtils(),
		viewCache:   template.New(""),
		viewPaths:   cfg.GetPaths(),
		cached:      cfg.GetCached(),
	}
}

// NewViewCacheWithTemplates creates a new view cache wrapping the templates.
func NewViewCacheWithTemplates(templates *template.Template) *ViewCache {
	return &ViewCache{
		viewFuncMap: viewUtils(),
		viewCache:   templates,
		cached:      true,
	}
}

// ViewCache is the cached views used in view results.
type ViewCache struct {
	viewFuncMap  template.FuncMap
	viewPaths    []string
	viewLiterals []string
	viewCache    *template.Template
	cached       bool
}

// SetCached sets if we should cache views once they're compiled, or always read them from disk.
// Cached == True, use in memory storage for views
// Cached == False, read the file from disk every time we want to render the view.
func (vc *ViewCache) SetCached(cached bool) {
	if vc == nil {
		return
	}
	vc.cached = cached
}

// Cached indicates if the cache is enabled, or if we skip parsing views each load.
// Cached == True, use in memory storage for views
// Cached == False, read the file from disk every time we want to render the view.
func (vc *ViewCache) Cached() bool {
	if vc == nil {
		return false
	}
	return vc.cached
}

// Initialize caches templates by path.
func (vc *ViewCache) Initialize() error {
	if len(vc.viewPaths) == 0 && len(vc.viewLiterals) == 0 {
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
func (vc *ViewCache) Parse() (views *template.Template, err error) {
	views = template.New("").Funcs(vc.viewFuncMap)
	if len(vc.viewPaths) > 0 {
		views, err = views.ParseFiles(vc.viewPaths...)
		if err != nil {
			err = exception.Wrap(err)
			return
		}
	}

	if len(vc.viewLiterals) > 0 {
		for _, viewLiteral := range vc.viewLiterals {
			views, err = views.Parse(viewLiteral)
			if err != nil {
				err = exception.Wrap(err)
				return
			}
		}
	}
	return
}

// AddPaths adds paths to the view collection.
func (vc *ViewCache) AddPaths(paths ...string) {
	if vc == nil {
		return
	}
	vc.viewPaths = append(vc.viewPaths, paths...)
}

// AddLiterals adds view literal strings to the view collection.
func (vc *ViewCache) AddLiterals(views ...string) {
	if vc == nil {
		return
	}
	vc.viewLiterals = append(vc.viewLiterals, views...)
}

// SetPaths sets the view paths outright.
func (vc *ViewCache) SetPaths(paths ...string) {
	if vc == nil {
		return
	}
	vc.viewPaths = paths
}

// SetLiterals sets the raw views outright.
func (vc *ViewCache) SetLiterals(viewLiterals ...string) {
	if vc == nil {
		return
	}
	vc.viewLiterals = viewLiterals
}

// Paths returns the view paths.
func (vc *ViewCache) Paths() []string {
	if vc == nil {
		return nil
	}
	return vc.viewPaths
}

// FuncMap returns the global view func map.
func (vc *ViewCache) FuncMap() template.FuncMap {
	if vc == nil {
		return nil
	}
	return vc.viewFuncMap
}

// Templates gets the view cache for the app.
func (vc *ViewCache) Templates() *template.Template {
	if vc == nil {
		return nil
	}
	return vc.viewCache
}

// SetTemplates sets the view cache for the app.
func (vc *ViewCache) SetTemplates(viewCache *template.Template) {
	if vc == nil {
		return
	}
	vc.viewCache = viewCache
}

func viewUtils() template.FuncMap {
	return template.FuncMap{
		"short": func(t time.Time) string {
			return t.Format("1/02/2006 3:04:05 PM")
		},
		"shortDate": func(t time.Time) string {
			return t.Format("1/02/2006")
		},
		"medium": func(t time.Time) string {
			return t.Format("Jan 02, 2006 3:04:05 PM")
		},
		"kitchen": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
		"monthDate": func(t time.Time) string {
			return t.Format("1/2")
		},
		"money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},
		"duration": func(d time.Duration) string {
			if d > time.Hour {
				return fmt.Sprintf("%0.2fh", float64(d)/float64(time.Hour))
			}
			if d > time.Minute {
				return fmt.Sprintf("%0.2fm", float64(d)/float64(time.Minute))
			}
			if d > time.Second {
				return fmt.Sprintf("%0.2fs", float64(d)/float64(time.Second))
			}
			if d > time.Millisecond {
				return fmt.Sprintf("%0.2fms", float64(d)/float64(time.Millisecond))
			}
			if d > time.Microsecond {
				return fmt.Sprintf("%0.2fµs", float64(d)/float64(time.Microsecond))
			}
			return fmt.Sprintf("%dns", d)
		},
		"pct": func(v float64) string {
			return fmt.Sprintf("%0.2f%%", v*100)
		},
		"csv": func(items []string) string {
			return strings.Join(items, ", ")
		},
		"json": func(v interface{}) (string, error) {
			contents, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(contents), nil
		},
		"jsonPretty": func(v interface{}) (string, error) {
			buf := bytes.NewBuffer(nil)
			encoder := json.NewEncoder(buf)
			encoder.SetIndent("", "\t")
			err := encoder.Encode(v)
			if err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}
}
