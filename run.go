package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

const MaxIdleConnsPerHost = 512

func run(ctx context.Context, cfg Config) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer cancel()

	clientProtocols := new(http.Protocols)
	clientProtocols.SetHTTP1(true)
	clientProtocols.SetHTTP2(false)

	httpTransport := &http.Transport{
		IdleConnTimeout:        90 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		DisableCompression:     true,
		Protocols:              clientProtocols,
		MaxConnsPerHost:        0, // no limit
		MaxResponseHeaderBytes: 10 << 20,
		MaxIdleConnsPerHost:    MaxIdleConnsPerHost,
		MaxIdleConns:           MaxIdleConnsPerHost * len(cfg.Websites),
	}

	hostToWebsite := make(map[string]Website)
	var domains []string
	for _, website := range cfg.Websites {
		hostToWebsite[website.Domain] = website
		domains = append(domains, website.Domain)
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(".cert-cache"),
		HostPolicy: autocert.HostWhitelist(domains...),
	}

	app := &application{
		httpClient:    &http.Client{Transport: httpTransport},
		hostToWebsite: hostToWebsite,
	}

	serverH1H2 := &http.Server{
		Handler:               app,
		DisableClientPriority: true,
		ErrorLog:              log.New(io.Discard, "", 0),
		TLSConfig: &tls.Config{
			GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
				website, ok := hostToWebsite[chi.ServerName]
				if !ok {
					return nil, fmt.Errorf("unknown domain %q", chi.ServerName)
				}

				var nextProtos []string
				if website.HTTP2 {
					nextProtos = append(nextProtos, "h2")
				}
				if website.HTTP1 {
					nextProtos = append(nextProtos, "http/1.1")
				}
				nextProtos = append(nextProtos, "acme-tls/1")

				return &tls.Config{
					GetCertificate: certManager.GetCertificate,
					NextProtos:     nextProtos,
				}, nil
			},
		},
	}

	ln80, err := net.Listen("tcp", ":80")
	if err != nil {
		return err
	}
	defer ln80.Close()
	log.Println("listening on :80")

	challengeServerProtocols := &http.Protocols{}
	challengeServerProtocols.SetHTTP1(true)
	challengeServerProtocols.SetHTTP2(false)
	challengeServer := &http.Server{
		Handler:   certManager.HTTPHandler(nil),
		ErrorLog:  log.New(io.Discard, "", 0),
		Protocols: challengeServerProtocols,
	}
	go func() {
		err := challengeServer.Serve(ln80)
		if err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	ln443, err := tls.Listen("tcp", ":443", serverH1H2.TLSConfig)
	if err != nil {
		return err
	}
	defer ln443.Close()
	log.Println("listening on :443")

	go func() {
		err := serverH1H2.Serve(ln443)
		if err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = challengeServer.Shutdown(shutdownCtx)
	_ = serverH1H2.Shutdown(shutdownCtx)
	return nil
}
