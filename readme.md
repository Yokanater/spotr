# spotr

Workout logging for nerds on the terminal

## Install

### Homebrew

```bash
brew tap Yokanater/tap
brew install --cask spotr
```

### Binaries

Download the archive for your system from the latest GitHub release:

- macOS Apple Silicon: `spotr_Darwin_arm64.tar.gz`
- macOS Intel: `spotr_Darwin_x86_64.tar.gz`
- Linux x86_64: `spotr_Linux_x86_64.tar.gz`
- Linux ARM64: `spotr_Linux_arm64.tar.gz`
- Windows x86_64: `spotr_Windows_x86_64.zip`
- Windows ARM64: `spotr_Windows_arm64.zip`

macOS/Linux:

```bash
tar -xzf spotr_*.tar.gz
./spotr
```

Windows PowerShell:

```powershell
Expand-Archive .\spotr_Windows_x86_64.zip
.\spotr_Windows_x86_64\spotr.exe
```

## Run

```bash
go run .
```

spotr writes to `spotr.db` in the current directory.

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
