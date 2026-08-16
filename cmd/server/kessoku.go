package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/mazrean/kessoku"
)

//go:generate go tool kessoku $GOFILE
var _ = kessoku.Inject[*App](
	"NewApp",

	// app
	kessoku.Provide(func() *App {
		return &App{
			httpsrv: &http.Server{Addr: "localhost:6060"},
		}
	}),
)
