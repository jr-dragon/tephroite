package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jr-dragon/tephroite/internal/repo/kv"
	"github.com/jr-dragon/tephroite/pkg/resp"
)

type Set struct {
	storage kv.KV
}

func (c *Set) String() string {
	return "SET"
}

func (c *Set) Exec(_ context.Context, args []resp.BulkString) (resp.Value, error) {
	p, err := c.parseArgs(args)
	if err != nil {
		switch {
		case errors.Is(err, errValidation):
			err := errWrongNumberOfArguments(c)
			return resp.NewSimpleError(err), fmt.Errorf("%w: %w", errValidation, err)
		}
	}

	// It doesn't failed right now, check it after implements list, ziplist or more.
	_ = c.storage.Set(p)

	return resp.OKValue, nil
}

func (c *Set) parseArgs(args []resp.BulkString) (kv.KVSetParam, error) {
	switch len(args) {
	case 3:
		return kv.NewKVSetParam(kv.NewKVKey(args[1]), kv.NewKVVal(args[2], nil)), nil
	default:
		return kv.KVSetParam{}, errValidation
	}
}
