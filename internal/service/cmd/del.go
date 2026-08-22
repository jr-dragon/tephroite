package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jr-dragon/tephroite/internal/repo/kv"
	"github.com/jr-dragon/tephroite/pkg/resp"
)

type Del struct {
	storage kv.KV
}

func (c *Del) String() string {
	return "DEL"
}

func (c *Del) Exec(_ context.Context, args []resp.BulkString) (resp.Value, error) {
	if len(args) < 2 {
		err := errWrongNumberOfArguments(c)
		return resp.NewSimpleError(err), fmt.Errorf("%w: %w", errValidation, err)
	}

	errs := make([]error, 0, len(args)-1)
	var deleted int
	for _, arg := range args[1:] {
		if err := c.storage.Del(kv.NewKVDelParam(kv.NewKVKey(arg))); err != nil {
			errs = append(errs, err)
		} else {
			deleted++
		}
	}

	return resp.Integer(deleted), errors.Join(errs...)
}
