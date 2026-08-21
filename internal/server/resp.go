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

	lnmu       sync.Mutex
	listener   net.Listener
	inShutdown atomic.Bool

	wg     sync.WaitGroup
	connmu sync.Mutex
	conns  map[net.Conn]struct{}
}

func NewRESPServer(handler *Handler) *RESPServer {
	return &RESPServer{
		Addr: ":16379",

		handler: handler,

		conns: make(map[net.Conn]struct{}),
	}
}

func (s *RESPServer) ListenAndServe() error {
	// Fast check if in shutting down state.
	if s.inShutdown.Load() {
		return ErrServerClosed
	}

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Track net.Listener in s.listener.
	s.lnmu.Lock()
	if s.inShutdown.Load() {
		s.lnmu.Unlock()
		_ = ln.Close()
		return ErrServerClosed
	}
	s.listener = ln
	s.lnmu.Unlock()

	// Untrack s.listener
	defer func() {
		s.lnmu.Lock()
		defer s.lnmu.Unlock()

		if s.listener == ln {
			s.listener = nil
		}
	}()

	for {
		conn, err := ln.Accept()
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
	s.connmu.Lock()
	defer s.connmu.Unlock()

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
		s.connmu.Lock()
		defer s.connmu.Unlock()

		delete(s.conns, conn)
	}()

	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	defer wr.Flush()

	cmd := resp.NewCommand(rd)

	for {
		ret, err := s.handler.ServeRESP(context.Background(), cmd)
		if err != nil {
			if s.inShutdown.Load() || errors.Is(err, io.EOF) {
				return
			}

			switch {
			case errors.Is(err, ErrClient):
				// do nothing
			case errors.Is(err, ErrClientFatal):
				wr.Write(ret.Marshal())
				return
			case errors.Is(err, ErrServer):
				slog.Error("failed to serve", slog.Any("error", err))
				wr.Write(ret.Marshal())
				return
			default:
				slog.Error("failed to serve", slog.Any("error", err))
				wr.Write(ret.Marshal())
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

	s.lnmu.Lock()
	ln := s.listener
	s.lnmu.Unlock()

	var errs []error
	if ln != nil {
		if err := ln.Close(); err != nil {
			errs = append(errs, err)
		}
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
	s.connmu.Lock()
	conns := maps.Clone(s.conns)
	s.connmu.Unlock()

	var errs []error
	for conn := range conns {
		errs = append(errs, conn.Close())
	}

	return errors.Join(errs...)
}
