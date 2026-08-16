# Tephroite Agent Guide

Tephroite is a high performance in-memory key-value storage service.

## Language and documentation

- Write all source code, test, comments, commit messages, and documentation in English.
- Keep every `AGENTS.md` under 1000 words. Move detailed guidance into the nearest package's `.agents/` directory and link to it from `AGENTS.md` when needed.

## Change discipline

- Keep changes scoped to the relevant package and do not modify unrelated user work.
- Update the relevant `README.md` when setup, commands, public behavior, or developer workflows change.
- After a code review, assess whether each reported finding requires a change. Apply warranted fixes directly; otherwise, explain why no change is needed. At the end of the review, list the report together with the fixes made or the rationale for not making changes.

## Pull requests

- Use GitHub for pull request workflows.
- Write the PR title in English using Conventional Commits, including appropriate scope when usefule. Example: `feat(kv): add key value service handler`
- Write the PR description in Traditional Chinese, format with `.github/pull_request_template.md`.
- Report commands actually run and their results in the verification section.
