package cmd

import (
	"context"

	"github.com/jr-dragon/tephroite/internal/repo/kv"
	"github.com/jr-dragon/tephroite/pkg/resp"
)

type Command interface {
	String() string
	Exec(context.Context, []resp.BulkString) (resp.Value, error)
}

func NewCommands(storage kv.KV) []Command {
	return []Command{
		&Ping{},

		&Set{storage: storage},
		&Get{storage: storage},
		&Del{storage: storage},
	}
}
