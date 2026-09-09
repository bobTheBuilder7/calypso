package main

import (
	"context"
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

	protocols := new(http.Protocols)

	protocols.SetHTTP1(true)
	protocols.SetHTTP2(false)

	httpTransport := &http.Transport{
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          MaxIdleConnsPerHost * len(cfg.Websites),
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
		Protocols:             protocols,
		MaxConnsPerHost:       0,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
	}

	app := application{
		httpClient: &http.Client{
			Transport: httpTransport,
		},
	}

	app.Start(ctx)
}
