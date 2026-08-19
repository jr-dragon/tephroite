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
