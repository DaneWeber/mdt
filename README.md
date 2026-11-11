# Markdown Pivot

Golang CLI tool that generates a blended Markdown file from multiple Markdown files.

## Example Input

```markdown
# Dwarves

## Day 3

- Mined for gold
  - Found a cave
  - Dislodged diamonds

## Day 4

- Continued mining with no success

# Elves

## Day 3

- Danced in the moonlight

## Day 4

- Wove crowns of flowers
```

## Example Output

```markdown
# Pivot

Generated at 2026-01-02T12:34:56Z

## Day 3

### Dwarves

- Mined for gold
  - Found a cave
  - Dislodged diamonds

### Elves

- Danced in the moonlight

## Day 4

### Dwarves

- Continued mining with no success

### Elves

- Wove crowns of flowers
```

## Building

Requires Go 1.26.5 or newer (see `go.mod`). The included DevContainer
(`.devcontainer/`) comes with a suitable Go toolchain already installed.

The `Makefile` wraps the common commands:

| Command        | What it does                                         |
| -------------- | ---------------------------------------------------- |
| `make`         | Build the `mdt` binary in the repository root        |
| `make install` | Install `mdt` to Go's bin directory                  |
| `make test`    | Run the tests                                        |
| `make clean`   | Remove the built binary                              |
| `make help`    | List the available targets                           |

The equivalent Go commands are below.

Build a binary in the repository root:

```bash
go build -o mdt .
./mdt --help
```

Or install it to `$(go env GOPATH)/bin` (or `$GOBIN`, if set) so `mdt` is
available everywhere:

```bash
go install .
```

If the shell then reports `mdt: command not found`, add Go's bin directory to
your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Run the tests:

```bash
go test ./...
```

## Usage

```bash
mdt pivot -o output.md input1.md input2.md
```

To take all Markdown files in the current directory and generate a pivot file:

```bash
mdt pivot -o output.md *.md
```
