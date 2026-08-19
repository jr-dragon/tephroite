package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	rd *bufio.Reader
}

func NewReader(rd io.Reader) *Reader {
	if buffered, ok := rd.(*bufio.Reader); ok {
		return &Reader{rd: buffered}
	}

	return &Reader{rd: bufio.NewReader(rd)}
}

func (rd *Reader) Read() (Value, error) {
	header, err := rd.rd.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(header) < len(SENTINEL)+1 || string(header[len(header)-len(SENTINEL):]) != SENTINEL {
		return nil, errors.New("invalid RESP header sentinel")
	}

	switch header[0] {
	case MAGIC_SIMPLE_STRING:
		return BuildSimpleString(header)
	case MAGIC_SIMPLE_ERROR:
		return BuildSimpleError(header)
	case MAGIC_INTEGER:
		return BuildInteger(header)
	case MAGIC_BULK_STRING:
		return BuildBulkString(header, rd.rd)
	case MAGIC_ARRAY:
		return BuildArray(header, rd.rd)
	}

	return nil, fmt.Errorf("unsupported RESP value type %q", header[0])
}
