package resp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

const (
	MAGIC_SIMPLE_STRING = '+'
	MAGIC_SIMPLE_ERROR  = '-'
	MAGIC_INTEGER       = ':'
	MAGIC_BULK_STRING   = '$'
	MAGIC_ARRAY         = '*'
	SENTINEL            = "\r\n"

	aggrBufferLowWatermark = 1 << 10
	aggrBufferGrowSize     = 4 << 10
	numericBufferMinSize   = 24
)

type Value interface {
	Marshal() []byte
	marshalTo(*bytes.Buffer)
}

type SimpleString []byte

func BuildSimpleString(src []byte) (SimpleString, error) {
	src = src[1 : len(src)-2]
	dst := make([]byte, len(src))
	copy(dst, src)
	return SimpleString(dst), nil
}

func (v SimpleString) marshalTo(buf *bytes.Buffer) {
	if buf == nil {
		buf = &bytes.Buffer{}
	}

	buf.WriteByte(MAGIC_SIMPLE_STRING)
	buf.Write(v)
	buf.WriteString(SENTINEL)
}

func (v SimpleString) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

const (
	ERRTYPE_DEFAULT = "ERR"
)

type SimpleError struct {
	typ string
	err error
}

func NewSimpleError(err error) SimpleError {
	return SimpleError{err: err, typ: ERRTYPE_DEFAULT}
}

func BuildSimpleError(src []byte) (SimpleError, error) {
	src = src[1 : len(src)-2]
	for i, b := range src {
		if b == ' ' {
			return SimpleError{typ: string(src[:i]), err: errors.New(string(src[i+1:]))}, nil
		}
	}

	return SimpleError{typ: string(src), err: nil}, nil
}

func (v SimpleError) marshalTo(buf *bytes.Buffer) {
	buf.WriteByte(MAGIC_SIMPLE_ERROR)
	buf.WriteString(v.typ)
	if v.err != nil {
		buf.WriteByte(' ')
		buf.WriteString(v.err.Error())
	}
	buf.WriteString(SENTINEL)
}

func (v SimpleError) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type Integer int64

func BuildInteger(src []byte) (Integer, error) {
	src = src[1 : len(src)-2]
	i, err := strconv.Atoi(string(src))
	return Integer(int64(i)), err
}

func (v Integer) marshalTo(buf *bytes.Buffer) {
	buf.WriteByte(MAGIC_INTEGER)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(v), 10))
	buf.WriteString(SENTINEL)
}

func (v Integer) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type BulkString struct {
	null bool
	data string
}

func NewBulkString(s string) BulkString {
	return BulkString{data: s}
}

func BuildBulkString(header []byte, rd io.Reader) (BulkString, error) {
	header = header[1 : len(header)-2]

	length, err := parseLength(header)
	if err != nil {
		if errors.Is(err, ErrNegLength) {
			return BulkString{null: true}, nil
		} else {
			return BulkString{}, nil
		}
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(rd, buf); err != nil {
		return BulkString{}, err
	}

	return BulkString{data: string(buf[:len(buf)-2])}, nil
}

func (v BulkString) marshalTo(buf *bytes.Buffer) {
	if v.null {
		buf.WriteString("$-1\r\n")
		return
	}
	if buf.Available() < len(v.data)+15 {
		buf.Grow(len(v.data) + 15)
	}

	buf.WriteByte(MAGIC_BULK_STRING)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)
	buf.WriteString(v.data)
	buf.WriteString(SENTINEL)
}

func (v BulkString) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

func (v BulkString) String() string {
	return v.data
}

type Array struct {
	null bool
	data []Value
}

func NewArray(data []Value) Array {
	return Array{data: data}
}

func BuildArray(header []byte, rd io.Reader) (Array, error) {
	header = header[1 : len(header)-2]
	if header[0] == '-' {
		return Array{null: true}, nil
	}
	values, err := readValues(header, rd)
	return Array{data: values}, err
}

func (v Array) marshalTo(buf *bytes.Buffer) {
	if v.null {
		buf.WriteString("*-1\r\n")
		return
	}

	growAggrBuffer(buf)

	buf.WriteByte(MAGIC_ARRAY)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)
	for _, v := range v.data {
		if v == nil {
			v = Null{}
		}
		v.marshalTo(buf)
	}
}

func (v Array) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}
