package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/jr-dragon/tephroite/internal/service/cmd"
	"github.com/jr-dragon/tephroite/pkg/resp"
)

var (
	ErrServer      = errors.New("handler: server error")
	ErrClient      = errors.New("handler: client error")
	ErrClientFatal = errors.New("handler: client fatal error")
)

type HandlerFunc func(context.Context, []resp.BulkString) (resp.Value, error)
type Middleware func(HandlerFunc) HandlerFunc

type Handler struct {
	executor map[string]HandlerFunc
	serve    HandlerFunc
}

func NewHandler(cmds []cmd.Command, middlewares ...Middleware) *Handler {
	h := Handler{}

	h.executor = make(map[string]HandlerFunc, len(cmds))
	for _, cmd := range cmds {
		h.executor[cmd.String()] = cmd.Exec
	}

	h.serve = h.Exec
	for _, m := range slices.Backward(middlewares) {
		h.serve = m(h.serve)
	}

	return &h
}

func (h *Handler) Exec(ctx context.Context, args []resp.BulkString) (resp.Value, error) {
	f, ok := h.executor[strings.ToUpper(args[0].String())]
	if !ok {
		err := fmt.Errorf("unknown command '%s'", args[0].String())
		return resp.NewSimpleError(err), fmt.Errorf("%w: %w", cmd.ErrCommand, err)
	}

	return f(ctx, args)
}

func (h *Handler) ServeRESP(ctx context.Context, rd io.Reader) (resp.Value, error) {
	args, err := resp.NewCommand(rd).Read()
	if err != nil {
		if errors.Is(err, resp.ErrProtocol) {
			return resp.NewSimpleError(err), fmt.Errorf("%w: %w", ErrClientFatal, err)
		} else {
			return resp.InternalError, fmt.Errorf("%w: %w", ErrServer, err)
		}
	}

	res, err := h.serve(ctx, args)
	if err != nil {
		if errors.Is(err, cmd.ErrCommand) {
			if res == nil {
				res = resp.UnknownError
			}
			return res, fmt.Errorf("%w: %w", ErrClient, err)
		} else {
			return resp.InternalError, fmt.Errorf("%w: %w", ErrServer, err)
		}
	}
	if res == nil {
		res = resp.OKValue
	}

	return res, nil
}
