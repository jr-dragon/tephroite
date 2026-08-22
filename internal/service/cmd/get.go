package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jr-dragon/tephroite/internal/repo/kv"
	"github.com/jr-dragon/tephroite/pkg/resp"
)

type Get struct {
	storage kv.KV
}

func (c *Get) String() string {
	return "GET"
}

func (c *Get) Exec(_ context.Context, args []resp.BulkString) (resp.Value, error) {
	if len(args) != 2 {
		err := errWrongNumberOfArguments(c)
		return resp.NewSimpleError(err), fmt.Errorf("%w: %w", errValidation, err)
	}

	v, err := c.storage.Get(kv.NewKVGetParam(kv.NewKVKey(args[1])))
	if err != nil {
		if errors.Is(err, kv.ErrKVNotFound) {
			return resp.Null{}, nil
		} else {
			// TODO: handle type mismatch or other errors
			return resp.NewSimpleError(err), err
		}
	}

	return resp.NewBulkString(v.String()), nil
}
