package resp

import "bytes"

var (
	OKValue = ok{}

	NullBulkStringValue = BulkString{null: true}
	NullArrayValue      = Array{null: true}
)

type ok struct{}

func (ok) marshalTo(buf *bytes.Buffer) { buf.WriteString("+OK\r\n") }

func (ok) Marshal() []byte { return []byte("+OK\r\n") }
