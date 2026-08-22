package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/mazrean/kessoku"

	"github.com/jr-dragon/tephroite/internal/repo/kv"
	"github.com/jr-dragon/tephroite/internal/server"
	"github.com/jr-dragon/tephroite/internal/service/cmd"
)

//go:generate go tool kessoku $GOFILE
var _ = kessoku.Inject[*App](
	"NewApp",
	// repository
	kessoku.Provide(kv.NewKV),

	// service
	kessoku.Provide(cmd.NewCommands),

	// server
	kessoku.Provide(func(cmds []cmd.Command) *server.Handler {
		return server.NewHandler(cmds)
	}),
	kessoku.Provide(server.NewRESPServer),

	// app
	kessoku.Provide(func(respsrv *server.RESPServer) *App {
		return &App{
			httpsrv: &http.Server{Addr: "localhost:6060"},
			respsrv: respsrv,
		}
	}),
)
