package resp

import (
	"bytes"
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

func (Null) Marshal() []byte {
	return []byte("_\r\n")
}

type Boolean bool

func (v Boolean) Marshal() []byte {
	if v {
		return []byte("#t\r\n")
	}

	return []byte("#f\r\n")
}

type Double float64

func (v Double) Marshal() []byte {
	buf := make([]byte, 0, 32)
	buf = append(buf, MAGIC_DOUBLE)

	switch value := float64(v); {
	case math.IsInf(value, 1):
		buf = append(buf, "inf"...)
	case math.IsInf(value, -1):
		buf = append(buf, "-inf"...)
	case math.IsNaN(value):
		buf = append(buf, "nan"...)
	default:
		buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
	}

	return append(buf, SENTINEL...)
}

type BigNumber big.Int

func NewBigNumber(value *big.Int) BigNumber {
	if value == nil {
		return BigNumber{}
	}

	return BigNumber(*new(big.Int).Set(value))
}

func (v BigNumber) Marshal() []byte {
	buf := make([]byte, 0, 64)
	buf = append(buf, MAGIC_BIG_NUMBER)
	buf = append(buf, (*big.Int)(&v).String()...)
	return append(buf, SENTINEL...)
}

type BulkError []byte

func NewBulkError(err error) BulkError {
	if err == nil {
		return nil
	}

	return BulkError(err.Error())
}

func (v BulkError) Marshal() []byte {
	header := make([]byte, 0, 64)
	header = append(header, MAGIC_BULK_ERROR)
	header = strconv.AppendInt(header, int64(len(v)), 10)
	header = append(header, SENTINEL...)

	buf := bytes.NewBuffer(header)
	buf.Write(v)
	buf.WriteString(SENTINEL)
	return buf.Bytes()
}

func (v BulkError) Error() string {
	return string(v)
}

type VerbatimString struct {
	encoding [3]byte
	data     string
}

func NewVerbatimString(encoding [3]byte, data string) VerbatimString {
	return VerbatimString{encoding: encoding, data: data}
}

func (v VerbatimString) Marshal() []byte {
	length := 4 + len(v.data)
	header := make([]byte, 0, 64)
	header = append(header, MAGIC_VERBATIM_STRING)
	header = strconv.AppendInt(header, int64(length), 10)
	header = append(header, SENTINEL...)

	buf := bytes.NewBuffer(header)
	buf.Write(v.encoding[:])
	buf.WriteByte(':')
	buf.WriteString(v.data)
	buf.WriteString(SENTINEL)
	return buf.Bytes()
}

type MapEntry struct {
	Key   Value
	Value Value
}

type Map struct {
	data []MapEntry
}

func NewMap(data []MapEntry) Map {
	return Map{data: data}
}

func (v Map) Marshal() []byte {
	buf := makeAggregateBuffer(len(v.data) * 2)
	buf = appendAggregateHeader(buf, MAGIC_MAP, len(v.data))
	for _, entry := range v.data {
		buf = appendValue(buf, entry.Key)
		buf = appendValue(buf, entry.Value)
	}
	return buf
}

type Attribute struct {
	data []MapEntry
}

func NewAttribute(data []MapEntry) Attribute {
	return Attribute{data: data}
}

func (v Attribute) Marshal() []byte {
	buf := makeAggregateBuffer(len(v.data) * 2)
	buf = appendAggregateHeader(buf, MAGIC_ATTRIBUTE, len(v.data))
	for _, entry := range v.data {
		buf = appendValue(buf, entry.Key)
		buf = appendValue(buf, entry.Value)
	}
	return buf
}

type Set struct {
	data []Value
}

func NewSet(data []Value) Set {
	return Set{data: data}
}

func (v Set) Marshal() []byte {
	return marshalAggregate(MAGIC_SET, v.data)
}

type Push struct {
	data []Value
}

func NewPush(data []Value) Push {
	return Push{data: data}
}

func (v Push) Marshal() []byte {
	return marshalAggregate(MAGIC_PUSH, v.data)
}

func marshalAggregate(magic byte, values []Value) []byte {
	buf := makeAggregateBuffer(len(values))
	buf = appendAggregateHeader(buf, magic, len(values))
	for _, value := range values {
		buf = appendValue(buf, value)
	}
	return buf
}

func makeAggregateBuffer(valueCount int) []byte {
	const (
		minimumCapacity         = 64
		averageEncodedValueSize = 16
	)

	return make([]byte, 0, minimumCapacity+valueCount*averageEncodedValueSize)
}

func appendAggregateHeader(buf []byte, magic byte, length int) []byte {
	buf = append(buf, magic)
	buf = strconv.AppendInt(buf, int64(length), 10)
	return append(buf, SENTINEL...)
}

func appendValue(buf []byte, value Value) []byte {
	if value == nil {
		return append(buf, MAGIC_NULL, '\r', '\n')
	}

	switch value := value.(type) {
	case SimpleString:
		buf = append(buf, MAGIC_SIMPLE_STRING)
		buf = append(buf, value...)
		return append(buf, SENTINEL...)
	case SimpleError:
		buf = append(buf, MAGIC_SIMPLE_ERROR)
		buf = append(buf, value.typ...)
		if value.err != nil {
			buf = append(buf, ' ')
			buf = append(buf, value.err.Error()...)
		}
		return append(buf, SENTINEL...)
	case Integer:
		buf = append(buf, MAGIC_INTEGER)
		buf = strconv.AppendInt(buf, int64(value), 10)
		return append(buf, SENTINEL...)
	case BulkString:
		if value.null {
			return append(buf, "$-1\r\n"...)
		}

		buf = append(buf, MAGIC_BULK_STRING)
		buf = strconv.AppendInt(buf, int64(len(value.data)), 10)
		buf = append(buf, SENTINEL...)
		buf = append(buf, value.data...)
		return append(buf, SENTINEL...)
	case Array:
		return appendArray(buf, value)
	case Null:
		return append(buf, MAGIC_NULL, '\r', '\n')
	case Boolean:
		buf = append(buf, MAGIC_BOOLEAN)
		if value {
			buf = append(buf, 't')
		} else {
			buf = append(buf, 'f')
		}
		return append(buf, SENTINEL...)
	case Double:
		buf = append(buf, MAGIC_DOUBLE)
		switch number := float64(value); {
		case math.IsInf(number, 1):
			buf = append(buf, "inf"...)
		case math.IsInf(number, -1):
			buf = append(buf, "-inf"...)
		case math.IsNaN(number):
			buf = append(buf, "nan"...)
		default:
			buf = strconv.AppendFloat(buf, number, 'g', -1, 64)
		}
		return append(buf, SENTINEL...)
	case BigNumber:
		buf = append(buf, MAGIC_BIG_NUMBER)
		buf = (*big.Int)(&value).Append(buf, 10)
		return append(buf, SENTINEL...)
	case BulkError:
		buf = append(buf, MAGIC_BULK_ERROR)
		buf = strconv.AppendInt(buf, int64(len(value)), 10)
		buf = append(buf, SENTINEL...)
		buf = append(buf, value...)
		return append(buf, SENTINEL...)
	case VerbatimString:
		buf = append(buf, MAGIC_VERBATIM_STRING)
		buf = strconv.AppendInt(buf, int64(4+len(value.data)), 10)
		buf = append(buf, SENTINEL...)
		buf = append(buf, value.encoding[:]...)
		buf = append(buf, ':')
		buf = append(buf, value.data...)
		return append(buf, SENTINEL...)
	case Map:
		return appendMapEntries(buf, MAGIC_MAP, value.data)
	case Attribute:
		return appendMapEntries(buf, MAGIC_ATTRIBUTE, value.data)
	case Set:
		return appendAggregateValues(buf, MAGIC_SET, value.data)
	case Push:
		return appendAggregateValues(buf, MAGIC_PUSH, value.data)
	default:
		return append(buf, value.Marshal()...)
	}
}

func appendArray(buf []byte, value Array) []byte {
	if value.null {
		return append(buf, "*-1\r\n"...)
	}

	buf = appendAggregateHeader(buf, MAGIC_ARRAY, len(value.data))
	for _, element := range value.data {
		if element == nil {
			buf = append(buf, "$-1\r\n"...)
			continue
		}
		buf = appendValue(buf, element)
	}
	return buf
}

func appendMapEntries(buf []byte, magic byte, entries []MapEntry) []byte {
	buf = appendAggregateHeader(buf, magic, len(entries))
	for _, entry := range entries {
		buf = appendValue(buf, entry.Key)
		buf = appendValue(buf, entry.Value)
	}
	return buf
}

func appendAggregateValues(buf []byte, magic byte, values []Value) []byte {
	buf = appendAggregateHeader(buf, magic, len(values))
	for _, value := range values {
		buf = appendValue(buf, value)
	}
	return buf
}
