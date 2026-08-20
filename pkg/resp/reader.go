package resp

import (
	"bufio"
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
		return nil, fmt.Errorf("resp: %w: %w", io.EOF, err)
	}
	if len(header) < 3 || header[len(header)-2] != '\r' {
		return nil, fmt.Errorf("%w: %s", errInvalidHeader, header)
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
	case MAGIC_NULL:
		return Null{}, nil
	case MAGIC_BOOLEAN:
		return BuildBoolean(header)
	case MAGIC_DOUBLE:
		return BuildDouble(header)
	case MAGIC_BIG_NUMBER:
		return BuildBigNumber(header)
	case MAGIC_BULK_ERROR:
		return BuildBulkError(header, rd.rd)
	case MAGIC_VERBATIM_STRING:
		return BuildVerbatimString(header, rd.rd)
	case MAGIC_MAP:
		return BuildMap(header, rd.rd)
	case MAGIC_ATTRIBUTE:
		return BuildAttribute(header, rd.rd)
	case MAGIC_SET:
		return BuildSet(header, rd.rd)
	case MAGIC_PUSH:
		return BuildPush(header, rd.rd)
	default:
		return BuildInlineArray(header)
	}
}
