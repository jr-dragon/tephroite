package resp

import (
	"bytes"
	"errors"
	"fmt"
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
	if err != nil {
		return Integer(int64(i)), fmt.Errorf("%w: failed to convert string to int: %s", errInvalidHeader, src)
	}
	return Integer(int64(i)), nil
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
	buf, err := readBlob(header, rd)
	if err != nil {
		if errors.Is(err, errNegativeLength) {
			return BulkString{null: true}, nil
		}
	}

	return BulkString{data: string(buf)}, err
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
	values, err := readValues(header, rd)
	if err != nil && errors.Is(err, errNegativeLength) {
		return Array{null: true}, nil
	}
	return Array{data: values}, err
}

func BuildInlineArray(header []byte) Array {
	splitted := bytes.Split(header[:len(header)-2], []byte{' '})

	strs := make([]Value, 0, len(splitted))
	for _, s := range splitted {
		strs = append(strs, NewBulkString(string(s)))
	}

	return Array{data: strs}
}

func BuildInlineBulkString(header []byte) []BulkString {
	splitted := bytes.Split(header[:len(header)-2], []byte{' '})

	strs := make([]BulkString, 0, len(splitted))
	for _, s := range splitted {
		strs = append(strs, NewBulkString(string(s)))
	}

	return strs
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
