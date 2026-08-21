# Tephroite

Tephroite is a high-performance, in-memory key-value storage service written in
Go. The current server is an early prototype that accepts RESP command arrays
and inline commands over TCP alongside a local pprof HTTP server.

## Requirements

- Go 1.26.6
- Make

Install the development tools used by the Make targets:

```sh
make init
```

This installs the latest `govulncheck` and `goimports` commands.

## Build and run

Generate the dependency-injection code and build the server:

```sh
make server
```

The executable is written to `bin/tp-server`. Start it with:

```sh
./bin/tp-server
```

The process starts these listeners:

| Service | Address | Current behavior |
| --- | --- | --- |
| RESP server | `tcp://:16379` | Executes the supported RESP or inline commands |
| pprof HTTP server | `http://localhost:6060/debug/pprof/` | Exposes Go runtime profiling endpoints locally |

The RESP listener binds to all available interfaces. The only implemented
command is `PING`, including its optional message argument. Commands may be
pipelined and receive responses in input order. For example:

```sh
printf 'PING\r\n*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n' | nc -w 1 localhost 16379
```

The server responds with `+PONG` followed by the bulk string `hello`. It does
not store data yet, so it is not a functional key-value server. See
[RESP support](./docs/resp.md) for command framing, supported wire values, and
current limitations.

Send `SIGINT` or `SIGTERM` to stop the process. The HTTP and RESP servers start
concurrently and share one lifecycle: if either server fails to start, the other
server is shut down. Coordinated shutdown has a one-second deadline. See
[server lifecycle](./docs/server-lifecycle.md) for synchronization details.

## Development

Run formatting and import organization:

```sh
make lint
```

Run code generation, vulnerability scanning, tests, and static analysis:

```sh
make test
```

Kessoku generates `cmd/server/kessoku_band.go` from
`cmd/server/kessoku.go`. Do not edit the generated file manually. Regenerate it
after changing dependency-injection providers:

```sh
go generate ./...
```

Remove build artifacts with:

```sh
make clean
```

## License

This software is under [MIT](./LICENSE) license.
