package resp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

var (
	ErrNegLength = errors.New("negative length")
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
	length, err := parseLength(header)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(rd, buf); err != nil {
		return nil, err
	}

	return buf[:len(buf)-2], nil
}

func readMapEntries(header []byte, rd io.Reader) ([]MapEntry, error) {
	length, err := parseLength(header)
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
	length, err := parseLength(header)
	if err != nil {
		return nil, err
	}

	values := make([]Value, 0, length)
	if length == 0 {
		return values, nil
	}
	if rd == nil {
		return values, io.ErrUnexpectedEOF
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

func parseLength(header []byte) (int, error) {
	if header[0] == '-' {
		return 0, ErrNegLength
	}

	return strconv.Atoi(string(header))
}
