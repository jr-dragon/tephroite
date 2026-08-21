# RESP support

Tephroite currently provides RESP wire-value parsing and marshaling in
`pkg/resp`, plus a command-oriented fast path used by the TCP server. The server
currently dispatches `PING` but does not operate a key-value store.

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

For the general value reader, an unrecognized first byte starts an inline
value. Inline values are split on ASCII spaces and represented as an array of
bulk strings. Quoting and escaping are not supported.

Streamed aggregates, streamed strings, and chunked strings are not supported.
Protocol negotiation is also not implemented; the reader determines the value
type from its prefix.

## Wire-value reader

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

## Command reader

The TCP server uses `resp.Command` instead of the general value reader. A
command can use either of these forms:

- An array with a positive length whose members are bulk strings.
- One inline line split on ASCII spaces.

For example, these inputs both represent `PING hello`:

```text
PING hello\r\n
```

```text
*2\r\n
$4\r\n
PING\r\n
$5\r\n
hello\r\n
```

`Command.Read` consumes exactly one command and leaves the next command for the
following call. Array members of any type other than bulk string are protocol
errors. Empty arrays, invalid lengths, missing CRLF sentinels, and incomplete
bulk-string bodies are also protocol errors.

The inline command format does not support quoting or escaping. Consecutive
spaces produce empty arguments.

## Supported commands

| Command | Arguments | Response |
| --- | --- | --- |
| `PING` | None | RESP simple string `PONG` |
| `PING` | One message | The message as a RESP bulk string |

Command names are matched case-insensitively. An unknown command or an invalid
argument count produces a RESP simple error and leaves the connection open for
the next command.

## TCP server behavior

The listener at `tcp://:16379` handles one connection per goroutine. It supports
pipelined commands and writes their responses in input order.

A protocol error produces a RESP simple error and closes the connection after
the response. EOF closes the connection without a response. An unexpected
internal handler error produces an `ERR internal server error` response and
closes the connection.

Only `PING` is implemented. Database state, authentication, transactions,
protocol negotiation, streamed aggregates, streamed strings, and chunked
strings are not supported.
