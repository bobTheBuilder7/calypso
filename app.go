package main

import (
	"context"
	"net/http"
)

type application struct {
	httpClient *http.Client
}

func (app *application) Start(ctx context.Context) error
