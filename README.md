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

## Usage

```bash
mdt pivot -o output.md input1.md input2.md
```

To take all Markdown files in the current directory and generate a pivot file:

```bash
mdt pivot -o output.md *.md
```
