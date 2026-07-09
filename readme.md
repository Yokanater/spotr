# ruffnut

Workout logging for nerds on the terminal

## Run

```bash
go run .
```

ruffnut writes to `ruffnut.db` in the current directory.

## Test

```bash
go test ./...
```

## Core Keys

- `j` / `k`: move
- `enter`: open selected item
- `a`: add
- `s`: start workout log
- `l`: log an exercise
- `v`: view logs
- `f`: finish workout log
- `e`: edit
- `d`: delete with confirmation
- `t`: browse templates
- `b` / `esc`: back
- `:`: command mode
- `?`: help
- `q`: quit with confirmation

Arrow keys still work as aliases

## Templates

Templates are JSON program definitions stored in `templates/programs/`.
Use them to start a program from a shared workout template, or export your own
program as a template file.

Press `t` in normal mode to browse templates with `j` / `k`, then press
`enter` to import the selected template as a program.

Set `SPOTR_TEMPLATE_DIR` to browse, import, export, and validate templates from
a custom directory.

Command mode supports:

```text
template list
template show <name|path>
template import <name|path>
template workout <template> <workout>
template export [program] [path]
template validate [name|path]
```

Examples:

```text
:template list
:template show Push Pull Legs
:template import Push Pull Legs
:template workout Push Pull Legs Push
:template export Push Pull Legs templates/programs/my-ppl.json
:template validate
```

Community templates can be added by opening a PR with a JSON file under
`templates/programs/`. See `templates/README.md` and
`templates/schema/program-template.schema.json`. PRs will run `go test ./...` in CI
