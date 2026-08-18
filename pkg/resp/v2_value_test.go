package resp

import (
	"errors"
	"math"
	"testing"
)

var benchmarkMarshalResult []byte

func TestSimpleStringMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value SimpleString
		want  string
	}{
		{
			name:  "empty",
			value: SimpleString(""),
			want:  "+\r\n",
		},
		{
			name:  "text",
			value: SimpleString("PONG"),
			want:  "+PONG\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestSimpleErrorMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value SimpleError
		want  string
	}{
		{
			name:  "nil error",
			value: NewSimpleError(nil),
			want:  "-ERR\r\n",
		},
		{
			name:  "error message",
			value: NewSimpleError(errors.New("unknown command")),
			want:  "-ERR unknown command\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestIntegerMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Integer
		want  string
	}{
		{
			name:  "zero",
			value: 0,
			want:  ":0\r\n",
		},
		{
			name:  "negative",
			value: -42,
			want:  ":-42\r\n",
		},
		{
			name:  "maximum",
			value: Integer(math.MaxInt64),
			want:  ":9223372036854775807\r\n",
		},
		{
			name:  "minimum",
			value: Integer(math.MinInt64),
			want:  ":-9223372036854775808\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestBulkStringMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value BulkString
		want  string
	}{
		{
			name:  "null",
			value: NullBulkStringValue,
			want:  "$-1\r\n",
		},
		{
			name:  "empty",
			value: NewBulkString(""),
			want:  "$0\r\n\r\n",
		},
		{
			name:  "text",
			value: NewBulkString("hello"),
			want:  "$5\r\nhello\r\n",
		},
		{
			name:  "length is measured in bytes",
			value: NewBulkString("你好"),
			want:  "$6\r\n你好\r\n",
		},
		{
			name:  "embedded sentinel",
			value: NewBulkString("a\r\nb"),
			want:  "$4\r\na\r\nb\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func TestBulkStringString(t *testing.T) {
	const want = "hello"

	if got := NewBulkString(want).String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestArrayMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Array
		want  string
	}{
		{
			name:  "null",
			value: NullArrayValue,
			want:  "*-1\r\n",
		},
		{
			name:  "empty",
			value: NewArray(nil),
			want:  "*0\r\n",
		},
		{
			name: "mixed values",
			value: NewArray([]Value{
				SimpleString("OK"),
				Integer(-2),
				NewBulkString("hey"),
				NewSimpleError(errors.New("failure")),
			}),
			want: "*4\r\n+OK\r\n:-2\r\n$3\r\nhey\r\n-ERR failure\r\n",
		},
		{
			name:  "static OK value",
			value: NewArray([]Value{OKValue}),
			want:  "*1\r\n+OK\r\n",
		},
		{
			name:  "nil element is null",
			value: NewArray([]Value{nil}),
			want:  "*1\r\n_\r\n",
		},
		{
			name:  "null bulk string",
			value: NewArray([]Value{NullBulkStringValue}),
			want:  "*1\r\n$-1\r\n",
		},
		{
			name: "nested array",
			value: NewArray([]Value{
				NewArray([]Value{NewBulkString("PING")}),
			}),
			want: "*1\r\n*1\r\n$4\r\nPING\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshaledValue(t, tt.value, tt.want)
		})
	}
}

func assertMarshaledValue(t *testing.T, value Value, want string) {
	t.Helper()

	if got := string(value.Marshal()); got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func BenchmarkSimpleStringMarshal(b *testing.B) {
	value := SimpleString("PONG")

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkSimpleErrorMarshal(b *testing.B) {
	value := NewSimpleError(errors.New("unknown command"))

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkIntegerMarshal(b *testing.B) {
	value := Integer(math.MaxInt64)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkBulkStringMarshal(b *testing.B) {
	value := NewBulkString("hello")

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}

func BenchmarkArrayMarshal(b *testing.B) {
	values := make([]Value, 100)
	for i := range values {
		values[i] = NewBulkString("hello")
	}
	value := NewArray(values)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkMarshalResult = value.Marshal()
	}
}
