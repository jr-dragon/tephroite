package server

import (
	"context"
	"log/slog"
	"strings"
	"sync"

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
	buf, _ := c.Next(-1)
	c.Write([]byte(strings.ToUpper(string(buf))))

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
