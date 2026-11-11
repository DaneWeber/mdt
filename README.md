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
mdt pivot [flags] input.md...
```

| Flag              | Meaning                                                              |
| ----------------- | -------------------------------------------------------------------- |
| `-o, --output`    | Output file (default: stdout). Overwritten if it exists.             |
| `-l, --levels N`  | Swap headings and lists: the first N list levels become headings,    |
|                   | and the heading levels become nested list items.                     |
| `-s, --skip N`    | With `--levels`: keep the top N heading levels as-is (default 1).    |

Multiple input files are concatenated before pivoting.

### Swapping headings and lists (`--levels`)

Given a log organized by date, with groups as list items:

```markdown
# Log

## Year 1

### Day 1

- Dwarves
  - Mined for gold
- Elves
  - Danced
```

`mdt pivot -l 1 log.md` turns the groups (1 list level) into headings and the
dates into a nested list:

```markdown
# Log

## Dwarves

- Year 1
  - Day 1
    - Mined for gold

## Elves

- Year 1
  - Day 1
    - Danced
```

Running `mdt pivot -l 2` on that output turns the two date levels back into
headings, restoring the original. HTML comments such as `<!-- NOW -->` under
a heading are carried along. Paragraphs under a pivoted heading can't be
placed in a list and are dropped, with a warning on stderr.

`--skip` controls how many top heading levels are left alone. The default of
1 keeps the document title. With `--skip 2`, each `##` section is pivoted
on its own.

For the sample files:

```bash
mdt pivot -l 1 -o sample-files/background-activity-by-group.dm.md sample-files/background-activity.dm.md
mdt pivot -l 3 -o sample-files/background-activity.dm.md sample-files/background-activity-by-group.dm.md
```

### Plain pivot (no `--levels`)

Without `--levels`, H1 sections containing H2 subsections are swapped, as in
the example at the top. This can merge one-file-per-group notes by date:

```bash
mdt pivot -o by-day.md dwarves.md elves.md
```
