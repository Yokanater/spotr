# Contributing to spotr

Thanks for helping improve spotr. Bug reports, focused feature proposals,
documentation fixes, and workout templates are welcome.

## Before opening a change

- Search existing issues and pull requests.
- Open an issue before starting a large behavioral or interface change.
- Keep pull requests focused on one problem.
- Never include a real `spotr.db` or other personal training data.

## Development

spotr requires Go 1.26 or newer.

```bash
go test ./...
go run . --data-dir "$(mktemp -d)"
```

Run `gofmt` on changed Go files. Add tests for behavior changes and ensure
`go test ./...` passes before opening a pull request.

The website lives in `web/`; its local setup and commands are documented in
`web/README.md`.

## Program templates

Follow `templates/README.md` and the JSON schema in `templates/schema/`.
Template-only pull requests are tested by the same CI workflow as code changes.

By contributing, you agree that your contribution is licensed under the MIT
License used by this repository.
