package main

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type application struct {
	httpClient    *http.Client
	hostToWebsite map[string]Website
}

func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if strings.HasPrefix(strings.ToLower(host), "www.") {
		location := req.URL.Clone()
		location.Scheme = "http"
		if req.TLS != nil {
			location.Scheme = "https"
		}
		location.Host = strings.TrimPrefix(req.Host, "www.")
		http.Redirect(w, req, location.String(), http.StatusPermanentRedirect)
		return
	}

	website, ok := app.hostToWebsite[host]
	if !ok {
		http.Error(w, "no such domain", http.StatusNotFound)
		return
	}

	url := req.URL.Clone()
	url.Scheme = "http"
	url.Host = net.JoinHostPort("localhost", strconv.Itoa(int(website.Port)))

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

	for key, values := range proxyResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(proxyResp.StatusCode)
	_, _ = io.Copy(w, proxyResp.Body)
}
