package resp

var (
	OKValue = ok{}

	NullBulkStringValue = BulkString{null: true}
	NullArrayValue      = Array{null: true}
)

type ok struct{}

func (ok) Marshal() []byte { return []byte("+OK\r\n") }
