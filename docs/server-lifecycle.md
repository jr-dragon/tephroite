# Server lifecycle

The executable starts the RESP TCP server and the local pprof HTTP server in an
`errgroup`. Both servers share the group context. A `SIGINT`, `SIGTERM`, startup
failure, or runtime listener failure cancels the group and initiates shutdown of
both servers.

Shutdown uses a one-second deadline. The HTTP server performs its standard
graceful shutdown and falls back to `Close` if the deadline expires. The RESP
server stops accepting connections, closes tracked connections, and waits for
their goroutines until the deadline expires.

## RESP listener synchronization

`RESPServer` creates a listener as a local value before publishing it on the
server. Listener publication and lookup are protected by `lnmu`, while
`inShutdown` provides an atomic, permanent shutdown state.

After creating the listener, `ListenAndServe` locks `lnmu` and checks
`inShutdown` again. If shutdown has already started, it closes the unpublished
listener and returns `ErrServerClosed`. Otherwise it publishes the listener and
enters the accept loop.

`Shutdown` first sets `inShutdown`, then locks `lnmu` to obtain the published
listener. This ordering covers both relevant races:

- If shutdown wins, a later listener is rejected during publication.
- If publication wins, shutdown finds and closes the listener, unblocking
  `Accept`.

The shutdown state is never reset, so a `RESPServer` must not be reused after
shutdown.

## Connection synchronization

Accepted connections are registered in a map protected by `connmu`. The same
critical section starts the connection goroutine and increments its wait group.
During shutdown, setting `inShutdown` prevents new registrations. Taking
`connmu` to snapshot the connection map also forms a barrier with any
registration that began immediately before shutdown.

Shutdown closes every connection in the snapshot and waits for all registered
connection goroutines. A connection removes itself from the map before its
goroutine completes.

Changes to this lifecycle require tests for shutdown both before and after
listener publication. Connection changes must also verify that no wait-group
increment can occur after shutdown begins waiting.
