package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"

	"github.com/jr-dragon/tephroite/pkg/resp"
	"github.com/panjf2000/gnet/v2"
)

type RESPServer struct {
	Addr    string
	Handler *Handler

	gnet.BuiltinEventEngine

	mu         sync.Mutex
	eng        gnet.Engine
	booted     bool
	inShutdown bool
}

func NewRESPServer(handler *Handler) *RESPServer {
	return &RESPServer{
		Addr:    "tcp://:16379",
		Handler: handler,
	}
}

func (srv *RESPServer) OnBoot(eng gnet.Engine) gnet.Action {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	slog.Info("starting tephroite server:", slog.String("address", srv.Addr))

	if srv.inShutdown {
		return gnet.Shutdown
	}

	srv.eng = eng
	srv.booted = true
	return gnet.None
}

func (srv *RESPServer) OnTraffic(c gnet.Conn) gnet.Action {
	rd := bufio.NewReader(c)
	wr := bufio.NewWriter(c)
	defer wr.Flush()
	for {
		res, err := srv.Handler.ServeRESP(context.TODO(), rd)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				return gnet.None
			case errors.Is(err, ErrClient):
				wr.Write(res.Marshal())
				return gnet.None
			case errors.Is(err, ErrServer):
				slog.Error("failed to serve", slog.Any("error", err))
				wr.Write(resp.InternalError.Marshal())
				return gnet.Close
			default:
				slog.Error("failed to serve", slog.Any("error", err))
				wr.Write(resp.InternalError.Marshal())
				return gnet.None
			}
		}
		wr.Write(res.Marshal())
	}
}

func (srv *RESPServer) Shutdown(ctx context.Context) error {
	srv.mu.Lock()
	srv.inShutdown = true

	if !srv.booted {
		srv.mu.Unlock()
		return nil
	}

	eng := srv.eng
	srv.mu.Unlock()
	return eng.Stop(ctx)
}
