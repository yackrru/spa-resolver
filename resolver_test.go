package resolver_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	resolver "github.com/yackrru/spa-resolver"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var client = new(http.Client)

func TestE2E(t *testing.T) {
	mux := http.NewServeMux()
	currentDir, _ := os.Getwd()

	setUpSpa(mux, currentDir)

	restRouting := map[string]string{
		"/foo": "Foo!",
		"/bar": "Bar!",
	}
	for k, v := range restRouting {
		path := k
		body := v
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, body)
		})
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	// Tests for SPA (single page)
	orgIndex, _ := ioutil.ReadFile(currentDir + "/testdata/index.html")
	spaRouting := []string{
		"/",
		"/baz",
		"/foo/bar",
		"/bar/baz",
		"/testdata/static/index.js",
		"/resolver.go",
		"/static/../../resolver.go",
	}
	for _, path := range spaRouting {
		req, _ := http.NewRequest("GET", server.URL+path, nil)
		res, _ := client.Do(req)
		body, _ := ioutil.ReadAll(res.Body)

		expected := string(orgIndex)
		assert.Equal(t, expected, string(body))
		assert.Equal(t, 200, res.StatusCode)
	}

	// Tests for SPA (resources)
	orgJs, _ := ioutil.ReadFile(currentDir + "/testdata/static/index.js")
	orgCSS, _ := ioutil.ReadFile(currentDir + "/testdata/assets/index.css")
	resourcesRouting := map[string]string{
		"/static/index.js":  string(orgJs),
		"/assets/index.css": string(orgCSS),
	}
	for path, expected := range resourcesRouting {
		req, _ := http.NewRequest("GET", server.URL+path, nil)
		res, _ := client.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.Equal(t, expected, string(body))
		assert.Equal(t, 200, res.StatusCode)
	}

	// Tests for SPA (not found resources)
	notExistsRouting := []string{"/static/foo.js"}
	for _, path := range notExistsRouting {
		req, _ := http.NewRequest("GET", server.URL+path, nil)
		res, _ := client.Do(req)
		assert.Equal(t, 404, res.StatusCode)
	}

	// Tests for REST
	for path, expected := range restRouting {
		req, _ := http.NewRequest("GET", server.URL+path, nil)
		res, _ := client.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.Equal(t, expected, string(body))
		assert.Equal(t, 200, res.StatusCode)
	}
}

func setUpSpa(mux *http.ServeMux, currentDir string) {
	config := resolver.NewSpaConfig(mux)

	resources := []resolver.Resource{
		{
			Dir:  currentDir + "/testdata/static",
			Path: "/static",
		},
		{
			Dir:  currentDir + "/testdata/assets",
			Path: "/assets",
		},
	}
	config.DefineResources(resources...)

	sp := &resolver.SinglePage{
		Dir:  currentDir + "/testdata",
		File: "index.html",
	}
	config.DefineSinglePage(sp)

	config.Build()
}

func TestSpPanic(t *testing.T) {
	config := resolver.NewSpaConfig(nil)

	defer func() {
		err := recover()
		if err != nil {
			assert.Equal(t, "Not found single page resource: ./index.html", err)
		}
	}()

	sp := &resolver.SinglePage{
		Dir:  ".",
		File: "index.html",
	}
	config.DefineSinglePage(sp)
}
