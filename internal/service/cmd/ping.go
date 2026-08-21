package cmd

import (
	"context"
	"fmt"

	"github.com/jr-dragon/tephroite/pkg/resp"
)

var pong = resp.SimpleString("PONG")

type Ping struct{}

func (c *Ping) String() string {
	return "PING"
}

func (c *Ping) Exec(_ context.Context, cmd []resp.BulkString) (resp.Value, error) {
	switch len(cmd) {
	case 1:
		return pong, nil
	case 2:
		return cmd[1], nil
	default:
		err := errWrongNumberOfArguments(c)
		return resp.NewSimpleError(err), fmt.Errorf("%w: %w", errValidation, err)
	}
}
