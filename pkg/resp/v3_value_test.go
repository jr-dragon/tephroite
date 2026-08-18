package resp

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"testing"
)

func TestNullMarshal(t *testing.T) {
	assertMarshaledValue(t, NullValue, "_\r\n")
}

func TestBooleanMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Boolean
		want  string
	}{
		{name: "true", value: true, want: "#t\r\n"},
		{name: "false", value: false, want: "#f\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestDoubleMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Double
		want  string
	}{
		{name: "zero", value: 0, want: ",0\r\n"},
		{name: "negative", value: -1.23, want: ",-1.23\r\n"},
		{name: "positive infinity", value: Double(math.Inf(1)), want: ",inf\r\n"},
		{name: "negative infinity", value: Double(math.Inf(-1)), want: ",-inf\r\n"},
		{name: "not a number", value: Double(math.NaN()), want: ",nan\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestBigNumberMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value *big.Int
		want  string
	}{
		{
			name:  "zero",
			value: big.NewInt(0),
			want:  "(0\r\n",
		},
		{
			name:  "nil is zero",
			value: nil,
			want:  "(0\r\n",
		},
		{
			name:  "larger than int64",
			value: mustBigInt(t, "3492890328409238509324850943850943825024385"),
			want:  "(3492890328409238509324850943850943825024385\r\n",
		},
		{
			name:  "negative",
			value: mustBigInt(t, "-3492890328409238509324850943850943825024385"),
			want:  "(-3492890328409238509324850943850943825024385\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, NewBigNumber(tt.value), tt.want)
		})
	}

	t.Run("zero value", func(t *testing.T) {
		assertMarshaledValue(t, BigNumber{}, "(0\r\n")
	})
}

func TestBulkErrorMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value BulkError
		want  string
	}{
		{name: "empty", value: BulkError(""), want: "!0\r\n\r\n"},
		{
			name:  "error",
			value: NewBulkError(errors.New("SYNTAX invalid syntax")),
			want:  "!21\r\nSYNTAX invalid syntax\r\n",
		},
		{name: "length is measured in bytes", value: BulkError("ERR 錯誤"), want: "!10\r\nERR 錯誤\r\n"},
		{name: "embedded sentinel", value: BulkError("ERR a\r\nb"), want: "!8\r\nERR a\r\nb\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestBulkErrorError(t *testing.T) {
	const want = "ERR failure"

	if got := BulkError(want).Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestVerbatimStringMarshal(t *testing.T) {
	textEncoding := [3]byte{'t', 'x', 't'}

	tests := []struct {
		name  string
		value VerbatimString
		want  string
	}{
		{
			name:  "text",
			value: NewVerbatimString(textEncoding, "Some string"),
			want:  "=15\r\ntxt:Some string\r\n",
		},
		{
			name:  "length is measured in bytes",
			value: NewVerbatimString(textEncoding, "你好"),
			want:  "=10\r\ntxt:你好\r\n",
		},
		{
			name:  "embedded sentinel",
			value: NewVerbatimString(textEncoding, "a\r\nb"),
			want:  "=8\r\ntxt:a\r\nb\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestMapMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Map
		want  string
	}{
		{name: "empty", value: NewMap(nil), want: "%0\r\n"},
		{
			name: "entries",
			value: NewMap([]MapEntry{
				{Key: SimpleString("first"), Value: Integer(1)},
				{Key: SimpleString("second"), Value: Integer(2)},
			}),
			want: "%2\r\n+first\r\n:1\r\n+second\r\n:2\r\n",
		},
		{
			name:  "nil key and value are null",
			value: NewMap([]MapEntry{{}}),
			want:  "%1\r\n_\r\n_\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestAttributeMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Attribute
		want  string
	}{
		{name: "empty", value: NewAttribute(nil), want: "|0\r\n"},
		{
			name: "entry",
			value: NewAttribute([]MapEntry{
				{Key: SimpleString("ttl"), Value: Integer(3600)},
			}),
			want: "|1\r\n+ttl\r\n:3600\r\n",
		},
		{
			name:  "nil key and value are null",
			value: NewAttribute([]MapEntry{{}}),
			want:  "|1\r\n_\r\n_\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestSetMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Set
		want  string
	}{
		{name: "empty", value: NewSet(nil), want: "~0\r\n"},
		{
			name:  "values",
			value: NewSet([]Value{SimpleString("orange"), Integer(42), nil}),
			want:  "~3\r\n+orange\r\n:42\r\n_\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestPushMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Push
		want  string
	}{
		{name: "empty", value: NewPush(nil), want: ">0\r\n"},
		{
			name: "event",
			value: NewPush([]Value{
				SimpleString("message"),
				NewBulkString("channel"),
				NewBulkString("payload"),
			}),
			want: ">3\r\n+message\r\n$7\r\nchannel\r\n$7\r\npayload\r\n",
		},
		{
			name:  "nil value is null",
			value: NewPush([]Value{nil}),
			want:  ">1\r\n_\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestAggregateMarshalMatchesValueMarshal(t *testing.T) {
	values := []Value{
		SimpleString("simple"),
		NewSimpleError(errors.New("failure")),
		Integer(-42),
		NewBulkString("bulk"),
		NewArray([]Value{Integer(1)}),
		NullValue,
		Boolean(true),
		Double(1.23),
		NewBigNumber(big.NewInt(1234567890)),
		BulkError("ERR failure"),
		NewVerbatimString([3]byte{'t', 'x', 't'}, "verbatim"),
		NewMap([]MapEntry{{Key: Integer(1), Value: Boolean(true)}}),
		NewAttribute([]MapEntry{{Key: Integer(2), Value: Double(2.5)}}),
		NewSet([]Value{Integer(3)}),
		NewPush([]Value{Integer(4)}),
	}

	want := marshalAggregateHeaderForTest(MAGIC_PUSH, len(values))
	for _, value := range values {
		want = append(want, value.Marshal()...)
	}
	assertMarshaledValue(t, NewPush(values), string(want))
}

func marshalAggregateHeaderForTest(magic byte, length int) []byte {
	return []byte(string(magic) + strconv.Itoa(length) + SENTINEL)
}

func mustBigInt(t testing.TB, value string) *big.Int {
	t.Helper()

	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		t.Fatalf("SetString(%q) failed", value)
	}
	return n
}

func BenchmarkNullMarshal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = NullValue.Marshal()
	}
}

func BenchmarkBooleanMarshal(b *testing.B) {
	value := Boolean(true)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkDoubleMarshal(b *testing.B) {
	value := Double(1.23)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkBigNumberMarshal(b *testing.B) {
	value := NewBigNumber(mustBigInt(b, "3492890328409238509324850943850943825024385"))

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkBulkErrorMarshal(b *testing.B) {
	value := BulkError("SYNTAX invalid syntax")

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkVerbatimStringMarshal(b *testing.B) {
	value := NewVerbatimString([3]byte{'t', 'x', 't'}, "Some string")

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkMapMarshal(b *testing.B) {
	entries := make([]MapEntry, 100)
	for i := range entries {
		entries[i] = MapEntry{Key: Integer(i), Value: NewBulkString("hello")}
	}
	value := NewMap(entries)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkAttributeMarshal(b *testing.B) {
	entries := make([]MapEntry, 100)
	for i := range entries {
		entries[i] = MapEntry{Key: Integer(i), Value: NewBulkString("hello")}
	}
	value := NewAttribute(entries)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkSetMarshal(b *testing.B) {
	values := make([]Value, 100)
	for i := range values {
		values[i] = Integer(i)
	}
	value := NewSet(values)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkPushMarshal(b *testing.B) {
	values := make([]Value, 100)
	for i := range values {
		values[i] = NewBulkString("hello")
	}
	value := NewPush(values)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}
