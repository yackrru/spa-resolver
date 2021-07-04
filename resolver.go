// resolver is a package that assists
// in the routing of the single page application.
package resolver

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Builder is the interface that wraps Build, DefineResources
// and DefineSinglePage methods.
// It is basically designed so that the single page application
// can be implemented by paying attention only to the Builder interface.
type Builder interface {

	// Build is the method that register the "/" routing
	// to http.HandleFunc or http.ServeMux.HandleFunc.
	// If the argument Mux of SpaConfig is nil, it will be executed
	// in the former, otherwise it will be executed in the latter.
	Build()

	// DefineResources is the method that configure static directory using Resource.
	// If the path defined here has http access, the application will return a static file,
	// and if the requested file does not exist, it will return a 404 http status code.
	DefineResources(resources ...Resource) Builder

	// DefineSinglePage is the method that configure so-called index.html of SPA.
	// Users can set paths other than "/" even when using the spa-resolver.
	// When using spa-resolver, a SinglePage file is returned for http access
	// other than any configuration paths, including static file settings.
	DefineSinglePage(page *SinglePage) Builder
}

// SpaConfig implements Builder and should not be generated directly externally,
// but with NewSpaConfig which returns a Builder interface.
type SpaConfig struct {
	// Paths of static files.
	Resources  []Resource

	SinglePage SinglePage

	// Mux is allowed to take nil.
	Mux        *http.ServeMux
}

var _ Builder = new(SpaConfig)

var global *SpaConfig

// Resource is the setting of static resources map.
// For Dir, specify the directory where the file is actually located,
// and for Path, specify which http path to listen for access to that directory.
type Resource struct {
	Dir  string
	Path string
}

// SinglePage is the setting of so-called index.html of SPA.
// For Dir, specify the directory where the file is located,
// and for File, specify file name like index.html.
type SinglePage struct {
	Dir  string
	File string
}

// NewSpaConfig generates a new SpaConfig instance.
// The value of mux is allowed to take nil.
func NewSpaConfig(mux *http.ServeMux) Builder {
	return &SpaConfig{
		Mux: mux,
	}
}

func (c *SpaConfig) Build() {
	global = c

	if c.Mux == nil {
		http.HandleFunc("/", handleSpa)
	} else {
		c.Mux.HandleFunc("/", handleSpa)
	}
}

func (c *SpaConfig) DefineResources(resources ...Resource) Builder {
	for _, r := range resources {
		c.Resources = append(c.Resources, r)
	}

	return c
}

// DefineSinglePage confirms the existence of the file
// and then adds the setting to SpaConfig.
// If the file does not exist, it will cause a panic.
func (c *SpaConfig) DefineSinglePage(page *SinglePage) Builder {
	if strings.HasSuffix(page.Dir, "/") {
		page.Dir = strings.TrimRight(page.Dir, "/")
	}

	if strings.HasPrefix(page.File, "/") {
		page.File = strings.TrimLeft(page.File, "/")
	}

	if _, err := os.Stat(page.String()); err != nil {
		panic("Not found single page resource: " + page.String())
	}

	c.SinglePage = *page

	return c
}

func (p *SinglePage) String() string {
	return p.Dir + "/" + p.File
}

func handleSpa(w http.ResponseWriter, r *http.Request) {
	config := global
	uri := r.URL.Path

	for _, rs := range config.Resources {
		if strings.HasPrefix(uri, rs.Path) {
			translated := strings.Replace(uri, rs.Path, rs.Dir, 1)

			b, err := ioutil.ReadFile(translated)
			if err != nil {
				w.WriteHeader(404)
				return
			}

			w.Write(b)
			return
		}
	}

	b, _ := ioutil.ReadFile(config.SinglePage.String())
	w.Write(b)
}
