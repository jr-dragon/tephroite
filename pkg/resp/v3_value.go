package resp

import (
	"bytes"
	"errors"
	"io"
	"math"
	"math/big"
	"strconv"
)

const (
	MAGIC_NULL            = '_'
	MAGIC_BOOLEAN         = '#'
	MAGIC_DOUBLE          = ','
	MAGIC_BIG_NUMBER      = '('
	MAGIC_BULK_ERROR      = '!'
	MAGIC_VERBATIM_STRING = '='
	MAGIC_MAP             = '%'
	MAGIC_ATTRIBUTE       = '|'
	MAGIC_SET             = '~'
	MAGIC_PUSH            = '>'
)

var NullValue = Null{}

type Null struct{}

func (v Null) marshalTo(buf *bytes.Buffer) {
	buf.WriteString("_\r\n")
}

func (v Null) Marshal() []byte {
	return []byte("_\r\n")
}

type Boolean bool

func (v Boolean) marshalTo(buf *bytes.Buffer) {
	if v {
		buf.WriteString("#t\r\n")
	} else {
		buf.WriteString("#f\r\n")
	}
}

func (v Boolean) Marshal() []byte {
	if v {
		return []byte("#t\r\n")
	} else {
		return []byte("#f\r\n")
	}
}

type Double float64

func BuildDouble(src []byte) (Double, error) {
	src = src[1 : len(src)-2]
	parsed, err := strconv.ParseFloat(string(src), 64)
	return Double(parsed), err
}

func (v Double) marshalTo(buf *bytes.Buffer) {
	buf.WriteByte(MAGIC_DOUBLE)

	switch val := float64(v); {
	case math.IsInf(val, 1):
		buf.WriteString("inf")
	case math.IsInf(val, -1):
		buf.WriteString("-inf")
	case math.IsNaN(val):
		buf.WriteString("nan")
	default:
		buf.Write(strconv.AppendFloat(numericBuffer(buf), val, 'g', -1, 64))
	}

	buf.WriteString(SENTINEL)
}

func (v Double) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type BigNumber struct {
	val *big.Int
}

func BuildBigNumber(src []byte) (BigNumber, error) {
	src = src[1 : len(src)-2]

	var ok bool
	n := BigNumber{}
	n.val, ok = new(big.Int).SetString(string(src), 10)
	if !ok {
		return BigNumber{}, errors.New("failed to parse big number")
	}
	return n, nil
}

func NewBigNumber(val *big.Int) BigNumber {
	if val == nil {
		return BigNumber{val: big.NewInt(0)}
	}

	return BigNumber{val: new(big.Int).Set(val)}
}

func (v BigNumber) marshalTo(buf *bytes.Buffer) {
	if v.val == nil {
		v.val = big.NewInt(0)
	}

	buf.WriteByte(MAGIC_BIG_NUMBER)
	buf.WriteString(v.val.String())
	buf.WriteString(SENTINEL)
}

func (v BigNumber) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type BulkError []byte

func BuildBulkError(header []byte, rd io.Reader) (BulkError, error) {
	data, err := readBlob(header, rd)
	if err != nil {
		return nil, err
	}

	return BulkError(data), nil
}

func NewBulkError(err error) BulkError {
	if err == nil {
		return nil
	}

	return BulkError(err.Error())
}

func (v BulkError) marshalTo(buf *bytes.Buffer) {
	if buf.Available() < len(v)+15 {
		buf.Grow(len(v) + 15)
	}

	buf.WriteByte(MAGIC_BULK_ERROR)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v)), 10))
	buf.WriteString(SENTINEL)
	buf.Write(v)
	buf.WriteString(SENTINEL)
}

func (v BulkError) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

func (v BulkError) Error() string {
	return string(v)
}

type VerbatimString struct {
	encoding [3]byte
	data     string
}

func BuildVerbatimString(header []byte, rd io.Reader) (VerbatimString, error) {
	payload, err := readBlob(header, rd)
	if err != nil {
		return VerbatimString{}, err
	}
	if len(payload) < 4 || payload[3] != ':' {
		return VerbatimString{}, errors.New("invalid verbatim string format")
	}

	var encoding [3]byte
	copy(encoding[:], payload[:3])

	return VerbatimString{encoding: encoding, data: string(payload[4:])}, nil
}

func NewVerbatimString(encoding [3]byte, data string) VerbatimString {
	return VerbatimString{encoding: encoding, data: data}
}

func (v VerbatimString) marshalTo(buf *bytes.Buffer) {
	if buf.Available() < len(v.data)+19 {
		buf.Grow(len(v.data) + 19)
	}
	buf.WriteByte(MAGIC_VERBATIM_STRING)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)+4), 10))
	buf.WriteString(SENTINEL)
	buf.Write(v.encoding[:])
	buf.WriteByte(':')
	buf.WriteString(v.data)
	buf.WriteString(SENTINEL)
}

func (v VerbatimString) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type MapEntry struct {
	Key   Value
	Value Value
}

type Map struct {
	data []MapEntry
}

func BuildMap(header []byte, rd io.Reader) (Map, error) {
	entries, err := readMapEntries(header, rd)
	return Map{data: entries}, err
}

func NewMap(data []MapEntry) Map {
	return Map{data: data}
}

func (v Map) marshalTo(buf *bytes.Buffer) {
	growAggrBuffer(buf)

	buf.WriteByte(MAGIC_MAP)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)

	for _, e := range v.data {
		if e.Key == nil {
			e.Key = Null{}
		}
		if e.Value == nil {
			e.Value = Null{}
		}

		e.Key.marshalTo(buf)
		e.Value.marshalTo(buf)
	}
}

func (v Map) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type Attribute struct {
	data []MapEntry
}

func BuildAttribute(header []byte, rd io.Reader) (Attribute, error) {
	entries, err := readMapEntries(header, rd)
	return Attribute{data: entries}, err
}

func NewAttribute(data []MapEntry) Attribute {
	return Attribute{data: data}
}

func (v Attribute) marshalTo(buf *bytes.Buffer) {
	growAggrBuffer(buf)

	buf.WriteByte(MAGIC_ATTRIBUTE)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)

	for _, e := range v.data {
		if e.Key == nil {
			e.Key = Null{}
		}
		if e.Value == nil {
			e.Value = Null{}
		}
		e.Key.marshalTo(buf)
		e.Value.marshalTo(buf)
	}
}

func (v Attribute) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type Set struct {
	data []Value
}

func BuildSet(header []byte, rd io.Reader) (Set, error) {
	values, err := readValues(header, rd)
	return Set{data: values}, err
}

func NewSet(data []Value) Set {
	return Set{data: data}
}

func (v Set) marshalTo(buf *bytes.Buffer) {
	growAggrBuffer(buf)

	buf.WriteByte(MAGIC_SET)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)
	for _, v := range v.data {
		if v == nil {
			v = Null{}
		}
		v.marshalTo(buf)
	}
}

func (v Set) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

type Push struct {
	data []Value
}

func BuildPush(header []byte, rd io.Reader) (Push, error) {
	values, err := readValues(header, rd)
	return Push{data: values}, err
}

func NewPush(data []Value) Push {
	return Push{data: data}
}

func (v Push) marshalTo(buf *bytes.Buffer) {
	growAggrBuffer(buf)

	buf.WriteByte(MAGIC_PUSH)
	buf.Write(strconv.AppendInt(numericBuffer(buf), int64(len(v.data)), 10))
	buf.WriteString(SENTINEL)
	for _, v := range v.data {
		if v == nil {
			v = Null{}
		}
		v.marshalTo(buf)
	}
}

func (v Push) Marshal() []byte {
	var buf bytes.Buffer
	v.marshalTo(&buf)
	return buf.Bytes()
}

func readBlob(header []byte, rd io.Reader) ([]byte, error) {
	length, err := parseLength(header)
	if err != nil {
		return nil, err
	}
	if rd == nil {
		return nil, io.ErrUnexpectedEOF
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(rd, data); err != nil {
		return nil, err
	}

	var sentinel [len(SENTINEL)]byte
	if _, err := io.ReadFull(rd, sentinel[:]); err != nil {
		return nil, err
	}
	if string(sentinel[:]) != SENTINEL {
		return nil, errors.New("invalid blob sentinel")
	}

	return data, nil
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
	length, err := strconv.Atoi(string(header[1 : len(header)-2]))
	if err != nil {
		return 0, err
	}
	if length < 0 {
		return 0, errors.New("length must not be negative")
	}

	return length, nil
}
