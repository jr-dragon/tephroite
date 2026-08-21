package cmd

import (
	"context"

	"github.com/jr-dragon/tephroite/pkg/resp"
)

type Command interface {
	String() string
	Exec(context.Context, []resp.BulkString) (resp.Value, error)
}

func NewCommands() []Command {
	return []Command{
		&Ping{},
	}
}
