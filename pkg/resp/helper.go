package resp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	aggrBufferLowWatermark = 1 << 10
	aggrBufferGrowSize     = 4 << 10
	numericBufferMinSize   = 24
)

var (
	ErrInvalidHeader      = errors.New("resp: invalid header")
	errNegativeLength     = errors.New("resp: negative length")
	errUnexpectedSentinel = errors.New("resp: unexpected sentinel")
)

// numericBuffer returns spare buffer storage for strconv's append functions.
// It reserves at least 24 bytes, enough for any base-10 int64 or float64 text.
func numericBuffer(buf *bytes.Buffer) []byte {
	if buf.Available() < numericBufferMinSize {
		buf.Grow(numericBufferMinSize)
	}

	return buf.AvailableBuffer()
}

// growAggrBuffer maintains a 1 KiB low-water mark for aggregate writes. When
// space falls below it, Grow reserves at least 4 KiB to amortize allocations.
func growAggrBuffer(buf *bytes.Buffer) {
	if buf.Available() < aggrBufferLowWatermark {
		buf.Grow(aggrBufferGrowSize)
	}
}

func readBlob(header []byte, rd io.Reader) ([]byte, error) {
	if len(header) < 3 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidHeader, header)
	}

	length, err := ParseLength(header[1 : len(header)-2])
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(rd, buf); err != nil {
		return nil, fmt.Errorf("resp: %w", io.ErrUnexpectedEOF)
	}
	if len(buf) < 2 || (buf[len(buf)-2] != '\r' && buf[len(buf)-1] != '\n') {
		return nil, errUnexpectedSentinel
	}

	return buf[:len(buf)-2], nil
}

func readMapEntries(header []byte, rd io.Reader) ([]MapEntry, error) {
	if len(header) < 3 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidHeader, header)
	}

	length, err := ParseLength(header[1 : len(header)-2])
	if err != nil {
		return nil, err
	}

	entries := make([]MapEntry, 0, length)
	if length == 0 {
		return entries, nil
	}
	if rd == nil {
		return entries, io.ErrUnexpectedEOF
	}

	vrd := NewReader(rd)
	for range length {
		key, err := vrd.Read()
		if err != nil {
			return entries, err
		}
		value, err := vrd.Read()
		if err != nil {
			return entries, err
		}
		entries = append(entries, MapEntry{Key: key, Value: value})
	}

	return entries, nil
}

func readValues(header []byte, rd io.Reader) ([]Value, error) {
	if len(header) < 3 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidHeader, header)
	}

	length, err := ParseLength(header[1 : len(header)-2])
	if err != nil {
		return nil, err
	}

	values := make([]Value, 0, length)
	if length == 0 {
		return values, nil
	}
	if rd == nil {
		return values, fmt.Errorf("resp: %w", io.ErrUnexpectedEOF)
	}

	vrd := NewReader(rd)
	for range length {
		value, err := vrd.Read()
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}

	return values, nil
}

func ParseLength(header []byte) (int, error) {
	if len(header) < 1 {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHeader, header)
	}

	if header[0] == '-' {
		return 0, errNegativeLength
	}

	return strconv.Atoi(string(header))
}
