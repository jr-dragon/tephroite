package resp

import (
	"fmt"
	"io"
)

const (
	// cmdArgsCacheCount is defined as 8 because the "SET" redis
	// command limits in 8 arguments.
	//
	// For variadic commands (eg. "MSET", "LPUSH", ...), it should
	// alllocs by make([]BulkString, 0, length).
	cmdArgsCacheCount = 8
)

type Command struct {
	rd *Reader

	// argsCache provides an array of command arguments to avoid
	// allocating by make([]BulkString, 0, length)
	argsCache [cmdArgsCacheCount]BulkString
}

func NewCommand(rd io.Reader) *Command {
	return &Command{rd: NewReader(rd)}
}

// Read is a fast path of reading []BulkString from Redis Commands.
//
// The (*Reader).Read() returns (Value, error), but Value is an
// interface which allocs memory on heap and costs on casting.
// This reciever returns ([]BulkString, error) directly instead.
func (c *Command) Read() ([]BulkString, error) {
	header, err := c.rd.readHeader()
	if err != nil {
		return nil, err
	}

	if header[0] != MAGIC_ARRAY {
		return BuildInlineBulkString(header), nil
	}

	sz, err := parseLength(header[1 : len(header)-2])
	if err != nil {
		return nil, err
	}
	if sz == 0 {
		return nil, fmt.Errorf("%w: array length must > 0 ", errInvalidHeader)
	}

	var args []BulkString
	if sz <= cmdArgsCacheCount {
		args = c.argsCache[:sz]
	} else {
		args = make([]BulkString, sz)
	}

	for i := range sz {
		h, err := c.rd.readHeader()
		if err != nil {
			return nil, err
		}

		arg, err := BuildBulkString(h, c.rd.rd)
		if err != nil {
			return nil, err
		}

		args[i] = arg
	}

	return args, nil
}
