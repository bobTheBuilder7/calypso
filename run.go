package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const MaxIdleConnsPerHost = 512

func run(ctx context.Context, cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer cancel()

	clientProtocols := new(http.Protocols)

	clientProtocols.SetHTTP1(true)
	clientProtocols.SetHTTP2(false)

	httpTransport := &http.Transport{
		IdleConnTimeout:        90 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		DisableCompression:     true,
		Protocols:              clientProtocols,
		MaxConnsPerHost:        0,        // no limit
		MaxResponseHeaderBytes: 10 << 20, // 10MB
		MaxIdleConnsPerHost:    MaxIdleConnsPerHost,
		MaxIdleConns:           MaxIdleConnsPerHost * len(cfg.Websites),
	}

	hostToPort := make(map[string]uint16)

	for _, website := range cfg.Websites {
		hostToPort[website.Domain] = website.Port
	}

	app := &application{
		httpClient: &http.Client{
			Transport: httpTransport,
		},
		hostToPort: hostToPort,
	}

	server := &http.Server{
		Handler:               app,
		DisableClientPriority: true,
		ErrorLog:              log.New(io.Discard, "", 0),
		TLSConfig: &tls.Config{
			GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
				return &tls.Config{}, nil
			},
		},
	}

	ln, err := net.Listen("tcp", ":80")
	if err != nil {
		panic(err.Error())
	}
	defer ln.Close()

	log.Println("listening on :80")

	go func() {
		err = server.Serve(ln)
		if err != nil {
			log.Println(err.Error())
		}
	}()

	<-ctx.Done()

	shutdownCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	server.Shutdown(shutdownCtx)

	return nil
}
