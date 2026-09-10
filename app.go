package main

import (
	"fmt"
	"net/http"
)

type application struct {
	httpClient *http.Client
	hostToPort map[string]uint16
}

func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	port, ok := app.hostToPort[req.Host]
	if !ok {
		http.Error(w, "no such domain", 500)
		return
	}

	url := req.URL.Clone()
	url.Host = fmt.Sprintf("%s:%d", "localhost", port)
	url.Scheme = "http"

	proxyReq, _ := http.NewRequestWithContext(req.Context(), req.Method, url.String(), req.Body)

	defer req.Body.Close()

	proxyReq.Header = req.Header

}
