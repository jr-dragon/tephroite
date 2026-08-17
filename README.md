# Tephroite

Tephroite is a high-performance, in-memory key-value storage service written in
Go. The current server is an early prototype that runs an uppercase echo service
on gnet alongside a local pprof HTTP server.

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
| Echo server | `tcp://:16379` | Converts each received payload to uppercase and writes it back |
| pprof HTTP server | `http://localhost:6060/debug/pprof/` | Exposes Go runtime profiling endpoints locally |

The echo listener binds to all available interfaces. It is a prototype and does
not yet implement the RESP protocol or key-value operations.

Send `SIGINT` or `SIGTERM` to stop the process. The HTTP and echo servers start
concurrently and share one lifecycle: if either server fails to start, the other
server is shut down. Graceful shutdown has a one-second deadline.

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
