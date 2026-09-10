package main

import (
	"io"
	"net"
	"net/http"
	"strconv"
)

type application struct {
	httpClient *http.Client
	hostToPort map[string]uint16
}

func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	port, ok := app.hostToPort[host]
	if !ok {
		http.Error(w, "no such domain", http.StatusNotFound)
		return
	}

	url := req.URL.Clone()
	url.Scheme = "http"
	url.Host = net.JoinHostPort("localhost", strconv.Itoa(int(port)))

	proxyReq, err := http.NewRequestWithContext(req.Context(), req.Method, url.String(), req.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	proxyReq.Header = req.Header.Clone()

	proxyResp, err := app.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer proxyResp.Body.Close()

	w.WriteHeader(proxyResp.StatusCode)

	for key, values := range proxyResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	_, _ = io.Copy(w, proxyResp.Body)
}
