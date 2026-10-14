package api

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

//go:embed docs.html
var docsPage []byte

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /docs", serveDocs)
	mux.HandleFunc("GET /docs/openapi.yaml", serveOpenAPI)
}

func serveDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(docsPage)
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Write(openAPISpec)
}
