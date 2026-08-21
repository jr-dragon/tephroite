package resp

import (
	"errors"
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

	// argsCache avoids allocating a new argument slice with make on every read.
	argsCache [cmdArgsCacheCount]BulkString

	// byteCache avoids allocating a new byte slice with make on every read.
	// Its size is 4 KiB minus the space occupied by rd and argsCache, keeping
	// the entire Command aligned to 4 KiB.
	byteCache [1<<12 - 200]byte
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

		if h[0] != MAGIC_BULK_STRING {
			return nil, fmt.Errorf("%w: expect '%c', got '%c'", errInvalidHeader, MAGIC_BULK_STRING, h[0])
		}
		arg, err := c.buildBulkString(h, c.rd.rd)
		if err != nil {
			return nil, err
		}

		args[i] = arg
	}

	return args, nil
}

func (c *Command) buildBulkString(header []byte, rd io.Reader) (BulkString, error) {
	if len(header) < 4 {
		return BulkString{}, fmt.Errorf("%w: %s", errInvalidHeader, header)
	}

	sz, err := parseLength(header[1 : len(header)-2])
	if err != nil {
		if errors.Is(err, errNegativeLength) {
			return NullBulkStringValue, nil
		}
		return BulkString{}, err
	}

	var buf []byte
	if sz+2 <= len(c.byteCache) {
		buf = c.byteCache[:sz+2]
	} else {
		buf = make([]byte, sz+2)
	}

	if _, err := io.ReadFull(rd, buf); err != nil {
		return BulkString{}, errUnexpectedEOF
	}
	if len(buf) < 2 || buf[len(buf)-2] != '\r' || buf[len(buf)-1] != '\n' {
		return BulkString{}, errUnexpectedSentinel
	}

	return BulkString{data: string(buf[:len(buf)-2])}, nil
}
