package resp

import (
	"bytes"
	"testing"
)

func TestReader_Read(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "simple string",
			payload: []byte("+OK\r\n"),
		},
		{
			name:    "simple error",
			payload: []byte("-ERR unknown error\r\n"),
		},
		{
			name:    "integer",
			payload: []byte(":12345\r\n"),
		},
		{
			name:    "bulk string",
			payload: []byte("$5\r\nhello\r\n"),
		},
		{
			name:    "array",
			payload: []byte("*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"),
		},
		{name: "null", payload: []byte("_\r\n")},
		{name: "boolean", payload: []byte("#t\r\n")},
		{name: "double", payload: []byte(",1.5\r\n")},
		{name: "big number", payload: []byte("(3492890328409238509324850943850943825024385\r\n")},
		{name: "bulk error", payload: []byte("!11\r\nERR failure\r\n")},
		{name: "verbatim string", payload: []byte("=8\r\ntxt:text\r\n")},
		{name: "map", payload: []byte("%1\r\n+key\r\n:1\r\n")},
		{name: "attribute", payload: []byte("|1\r\n+ttl\r\n:10\r\n")},
		{name: "set", payload: []byte("~2\r\n:1\r\n:2\r\n")},
		{name: "push", payload: []byte(">2\r\n+message\r\n$4\r\ndata\r\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReader(bytes.NewBuffer(tt.payload)).Read()
			if err != nil {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}

			assertMarshaledValue(t, got, string(tt.payload))
		})
	}
}
