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

- The gnet echo server and pprof HTTP server start concurrently and must share one lifecycle.
- Preserve the behavior that a startup or runtime error from either server cancels the group and shuts down both servers.
- Keep shutdown safe before and after gnet's `OnBoot`; changes to this synchronization require tests for both orderings.
- The current public listeners are the uppercase echo service on `tcp://:16379` and local pprof on `localhost:6060`.
- The echo service is not yet a RESP implementation. Do not document it as a functional key-value or RESP server until that protocol is implemented.

## Pull requests

- Use GitHub for pull request workflows.
- Write the PR title in English using Conventional Commits, including an appropriate scope when useful. Example: `feat(kv): add key value service handler`.
- Write the PR description in Traditional Chinese, format with `.github/pull_request_template.md`.
- Report commands actually run and their results in the verification section.
