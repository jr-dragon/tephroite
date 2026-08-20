package server

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jr-dragon/tephroite/pkg/resp"
	"github.com/panjf2000/gnet/v2"
)

type RESPServer struct {
	Addr string

	gnet.BuiltinEventEngine

	mu         sync.Mutex
	eng        gnet.Engine
	booted     bool
	inShutdown bool
}

func NewRESPServer() *RESPServer {
	return &RESPServer{
		Addr: "tcp://:16379",
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
	v, err := resp.NewReader(c).Read()
	if err != nil {
		c.Write(resp.NewSimpleError(err).Marshal())
		return gnet.None
	}

	slog.Info("recieved resp", slog.Any("value", v))
	c.Write(resp.OKValue.Marshal())

	return gnet.None
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
