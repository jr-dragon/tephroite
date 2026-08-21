# RESP support

Tephroite currently provides RESP wire-value parsing and marshaling in
`pkg/resp`. The server uses this package to validate incoming values, but it
does not dispatch commands or operate a key-value store yet.

## Supported values

| Protocol | Prefix | Value |
| --- | --- | --- |
| RESP2 | `+` | Simple string |
| RESP2 | `-` | Simple error |
| RESP2 | `:` | Integer |
| RESP2 | `$` | Bulk string, including the null form |
| RESP2 | `*` | Array, including the null form |
| RESP3 | `_` | Null |
| RESP3 | `#` | Boolean |
| RESP3 | `,` | Double |
| RESP3 | `(` | Big number |
| RESP3 | `!` | Bulk error |
| RESP3 | `=` | Verbatim string |
| RESP3 | `%` | Map |
| RESP3 | `|` | Attribute |
| RESP3 | `~` | Set |
| RESP3 | `>` | Push |

An unrecognized first byte is treated as the start of an inline value. Inline
values are split on ASCII spaces and represented as an array of bulk strings.
Quoting and escaping are not supported.

Streamed aggregates, streamed strings, and chunked strings are not supported.
Protocol negotiation is also not implemented; the reader determines the value
type from its prefix.

## Package contract

Create a reader around any `io.Reader` and call `Read` once per wire value:

```go
reader := resp.NewReader(source)
value, err := reader.Read()
```

`Read` consumes the complete value, including nested aggregate members, and
leaves the next value for the following call. Every parsed value implements
`resp.Value`; call `Marshal` to encode it with RESP framing:

```go
encoded := value.Marshal()
```

Blob and aggregate lengths are byte counts, and values use CRLF framing. Nil
members in arrays, maps, attributes, sets, and pushes marshal as RESP3 null
values.

## Server behavior

The listener at `tcp://:16379` reads all complete values currently available on
a connection. It buffers output and writes one RESP2 `+OK` response for each
successfully decoded value, preserving pipeline order. A malformed complete
value produces a RESP2 simple error response.

The decoded values are currently discarded. There is no command validation,
command execution, database state, authentication, or transaction support.
