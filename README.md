# SPA Resolver
[![CircleCI](https://circleci.com/gh/yackrru/spa-resolver/tree/main.svg?style=svg)](https://circleci.com/gh/yackrru/spa-resolver/tree/main)
[![CodeQL](https://github.com/ttksm/spa-resolver/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ttksm/spa-resolver/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ttksm/spa-resolver)](https://goreportcard.com/report/github.com/ttksm/spa-resolver)
[![Go Reference](https://pkg.go.dev/badge/github.com/ttksm/spa-resolver.svg)](https://pkg.go.dev/github.com/ttksm/spa-resolver)

Go library for resolving single page application paths.

## Overview
This library makes it easy to control:

- All accesses other than the explicitly set path will return index.html.
  - The `index.html` as the single page of SPA can be set in any directory and any file name.
- Can easily map the http path to the directory where static files such as JavaScript and CSS are located.
  - If the requested file does not exist, it will return 404 http status code.

## Install
```bash
go get -u github.com/ttksm/spa-resolver
```

## Examples

### (a) mux

```go
func main() {
    mux := http.NewServeMux()
    currentDir, _ := os.Getwd()

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
    config.DefineSinglePage(sp).Build()
    
    // Path as a REST API
    mux.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(writer, "Foo!")
    })

    http.ListenAndServe("127.0.0.1:8080", mux)
}
```

### (b) without mux

```go
func main() {
    config := resolver.NewSpaConfig(nil)
    currentDir, _ := os.Getwd()

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
    config.DefineSinglePage(sp).Build()

    // Path as a REST API
    http.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(writer, "Foo!")
    })

    server := http.Server{Addr: "127.0.0.1:8080"}
    server.ListenAndServe()
}
```

### (c) using [gorilla/mux](https://github.com/gorilla/mux)
- In this case, define the path on outside of resolver package without calling `Build()`.
- Note that the "/" setting must be last.

```go
func main() {
    r := mux.NewRouter()
    currentDir, _ := os.Getwd()
 
    restRouting := map[string]string{
        "/foo": "Foo!",
        "/bar": "Bar!",
    }
    for k, v := range restRouting {
        path := k
        body := v
        r.Path(path).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
            fmt.Fprintf(writer, body)
        })
    }

    config := resolver.NewSpaConfig(nil)

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

    resolver.Globalize(config)
    r.PathPrefix("/").HandlerFunc(resolver.HandleSpa)
    
    http.ListenAndServe("127.0.0.1:8080", r)
}
```

## License

MIT License
