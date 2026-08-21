# Tephroite Agent Guide

Tephroite is a high performance in-memory key-value storage service.

## Language and documentation

- Write all source code, test, comments, commit messages, and documentation in English.
- Keep every `AGENTS.md` under 1000 words. Move detailed guidance into the nearest package's `.agents/` directory and link to it from `AGENTS.md` when needed.

## Change discipline

- Keep changes scoped to the relevant package and do not modify unrelated user work.
- Update the relevant `README.md` when setup, commands, public behavior, or developer workflows change.
- After a code review, assess whether each reported finding requires a change. Apply warranted fixes directly; otherwise, explain why no change is needed. At the end of the review, list the report together with the fixes made or the rationale for not making changes.

## Development workflow

- Use Go 1.26.6 or the version declared by `go.mod` if it changes.
- Run `make init` to install `govulncheck` and `goimports` before using the development targets.
- Run `make lint` after editing Go source. It applies `goimports` and intentionally excludes Kessoku-generated files.
- Run `make test` before handing off a change. It regenerates DI code, runs `govulncheck`, executes Go tests, and runs `go vet`.
- Run `make server` to generate DI code and build `bin/tp-server`.

## Generated code

- Define Kessoku providers and injectors in `cmd/server/kessoku.go`.
- Treat `cmd/server/kessoku_band.go` as generated output. Do not edit or format it manually.
- Run `go generate ./...` after changing a provider or injector, and commit the resulting generated output.

## Server lifecycle

- The TCP RESP server and pprof HTTP server start concurrently and must share one lifecycle.
- Preserve the behavior that a startup or runtime error from either server cancels the group and shuts down both servers.
- Keep shutdown safe before and after listener publication; changes to listener or connection synchronization require tests for both orderings.
- The current public listeners are the experimental RESP service on `tcp://:16379` and local pprof on `localhost:6060`.
- A server is not reusable after shutdown. Keep listener publication and shutdown state synchronized.
- See [server lifecycle](docs/server-lifecycle.md) for the ownership and shutdown sequence.

## RESP implementation

- Keep protocol types and parsing in `pkg/resp`, server handling in `internal/server`, and command implementations in `internal/service/cmd`.
- The experimental server executes `PING` but does not store data. Keep public documentation explicit about the supported commands.
- Preserve ordered responses for pipelined commands. Return a RESP simple error for malformed non-EOF input and then close the connection.
- Keep command errors, such as unknown commands or invalid argument counts, non-fatal to the connection.
- A reader call must consume exactly one value, including nested aggregate content, without consuming the following value. Cover changes to aggregate parsing with a trailing-value test.
- A command reader call must consume exactly one command without consuming the following command. RESP array commands must contain bulk-string arguments.
- Keep wire lengths byte-based and require CRLF framing. See [RESP support](docs/resp.md) for the supported types, API contract, and known limitations.

## Pull requests

- Use GitHub for pull request workflows.
- Write the PR title in English using Conventional Commits, including an appropriate scope when useful. Example: `feat(kv): add key value service handler`.
- Write the PR description in Traditional Chinese, format with `.github/pull_request_template.md`.
- Report commands actually run and their results in the verification section.
