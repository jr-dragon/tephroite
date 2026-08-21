package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"maps"
	"net"
	"sync"
	"sync/atomic"

	"github.com/jr-dragon/tephroite/pkg/resp"
)

var (
	ErrServerClosed = errors.New("resp: server closed")
)

type RESPServer struct {
	Addr string

	handler *Handler

	listener   net.Listener
	inShutdown atomic.Bool

	wg    sync.WaitGroup
	mu    sync.Mutex
	conns map[net.Conn]struct{}
}

func NewRESPServer(handler *Handler) *RESPServer {
	return &RESPServer{
		handler: handler,

		conns: make(map[net.Conn]struct{}),
	}
}

func (s *RESPServer) ListenAndServe() error {
	var err error
	if s.listener, err = net.Listen("tcp", ":16379"); err != nil {
		return err
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.inShutdown.Load() {
				return ErrServerClosed
			}

			slog.Error("failed to accept connection", slog.Any("error", err))
			return err
		}

		if !s.trackConn(conn) {
			_ = conn.Close()
			return ErrServerClosed
		}
	}
}

func (s *RESPServer) trackConn(conn net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inShutdown.Load() {
		return false
	}

	s.conns[conn] = struct{}{}
	s.wg.Go(func() { s.serve(conn) })
	return true
}

func (s *RESPServer) serve(conn net.Conn) {
	defer conn.Close()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.conns, conn)
	}()

	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	defer wr.Flush()

	for {
		ret, err := s.handler.ServeRESP(context.Background(), rd)
		if err != nil {
			if s.inShutdown.Load() || errors.Is(err, io.EOF) {
				return
			}

			switch {
			case errors.Is(err, ErrClient):
				// do nothing
			case errors.Is(err, ErrClientFatal):
				return
			case errors.Is(err, ErrServer):
				slog.Error("failed to serve", slog.Any("error", err))
				return
			default:
				slog.Error("failed to serve", slog.Any("error", err))
				return
			}
		}

		if _, err := wr.Write(ret.Marshal()); err != nil {
			if s.inShutdown.Load() {
				return
			}

			slog.Error("failed to write to conn:", slog.Any("error", err))
			return
		}
		if errors.Is(err, resp.ErrProtocol) {
			return
		}

		if rd.Buffered() == 0 {
			if err := wr.Flush(); err != nil {
				if s.inShutdown.Load() {
					return
				}

				slog.Error("failed to flush", slog.Any("error", err))
				return
			}
		}
	}
}

func (s *RESPServer) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)
	if s.listener == nil {
		return nil
	}

	var errs []error
	if err := s.listener.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := s.closeConns(); err != nil {
		errs = append(errs, err)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return errors.Join(errs...)
	case <-ctx.Done():
		errs = append(errs, ctx.Err())
		return errors.Join(errs...)
	}
}

func (s *RESPServer) closeConns() error {
	s.mu.Lock()
	conns := maps.Clone(s.conns)
	s.mu.Unlock()

	var errs []error
	for conn := range conns {
		errs = append(errs, conn.Close())
	}

	return errors.Join(errs...)
}
