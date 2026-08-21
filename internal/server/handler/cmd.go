package handler

import (
	"io"

	"github.com/jr-dragon/tephroite/pkg/resp"
)

const (
	// cmdArgsCacheCount is defined as 8 because the "SET" redis
	// command limits in 8 arguments.
	//
	// For variadic commands (eg. "MSET", "LPUSH", ...), it should
	// alllocs by make([]resp.BulkString, 0, length).
	cmdArgsCacheCount = 8
)

type command struct {
	*resp.Reader

	// argsCache provides an array of command arguments to avoid
	// allocating by make([]resp.BulkString, 0, length)
	argsCache [cmdArgsCacheCount]resp.BulkString
}

func newCommand(rd io.Reader) *command {
	return &command{Reader: resp.NewReader(rd)}
}

// Read is a fast path of reading []resp.BulkString from Redis Commands.
//
// The resp.(*Reader).Read() returns (resp.Value, error), but resp.Value
// is an interface which allocs memory on heap and costs on casting.
// This reciever returns ([]resp.BulkString, error) directly instead.
func (c *command) Read() ([]resp.BulkString, error) {
	header, err := c.Reader.ReadHeader()
	if err != nil {
		return nil, err
	}

	length, err := resp.ParseLength(header[1 : len(header)-2])
	if err != nil {
		return nil, err
	}

	var args []resp.BulkString
	if length <= cmdArgsCacheCount {
		args = c.argsCache[:cmdArgsCacheCount]
	} else {
		args = make([]resp.BulkString, 0, length)
	}

	for range length {
		h, err := c.Reader.ReadHeader()
		if err != nil {
			return nil, err
		}

		arg, err := resp.BuildBulkString(h, c.Reader.Reader())
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	return args, nil
}
