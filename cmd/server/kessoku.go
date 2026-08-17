package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/mazrean/kessoku"

	"github.com/jr-dragon/tephroite/internal/server"
)

//go:generate go tool kessoku $GOFILE
var _ = kessoku.Inject[*App](
	"NewApp",

	// server
	kessoku.Provide(server.NewRESPServer),

	// app
	kessoku.Provide(func(respsrv *server.RESPServer) *App {
		return &App{
			httpsrv: &http.Server{Addr: "localhost:6060"},
			respsrv: respsrv,
		}
	}),
)
