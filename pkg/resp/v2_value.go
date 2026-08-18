package resp

import (
	"bytes"
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

type Value interface {
	Marshal() []byte
	marshalTo(*bytes.Buffer)
}

type SimpleString []byte

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
