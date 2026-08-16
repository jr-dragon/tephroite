package server

import (
	"context"
	"log/slog"
	"strings"

	"github.com/panjf2000/gnet/v2"
)

type RESPServer struct {
	Addr string

	gnet.BuiltinEventEngine
	eng gnet.Engine
}

func NewRESPServer() *RESPServer {
	return &RESPServer{
		Addr: "tcp://:16379",
	}
}

func (srv *RESPServer) OnBoot(eng gnet.Engine) gnet.Action {
	slog.Info("starting tephroite server:", slog.String("address", srv.Addr))

	srv.eng = eng
	return gnet.None
}

func (srv *RESPServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	c.Write([]byte(strings.ToUpper(string(buf))))

	return gnet.None
}

func (srv *RESPServer) Shutdown(ctx context.Context) error {
	return srv.eng.Stop(ctx)
}
