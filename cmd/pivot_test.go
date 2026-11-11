package cmd

import (
	"bytes"
	"strings"
	"testing"
)

const pivotInput = `# Dwarves

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
`

const pivotExpected = `## Day 3

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
`

func TestPivot(t *testing.T) {
	input := strings.NewReader(pivotInput)
	var output bytes.Buffer

	err := pivot(input, &output)
	if err != nil {
		t.Fatalf("pivot returned error: %v", err)
	}

	got := output.String()
	if got != pivotExpected {
		t.Errorf("pivot output mismatch\ngot:\n%s\nwant:\n%s", got, pivotExpected)
	}
}

const byGroupInput = `# Log

## Year 1

### Day 1

- Dwarves
  - Mined for gold
    - Found a cave
- Elves

### Day 2

<!-- NOW -->

- Dwarves
- Elves
  - Danced
`

const byGroupExpected = `# Log

## Dwarves

- Year 1
  - Day 1
    - Mined for gold
      - Found a cave
  - Day 2
    <!-- NOW -->

## Elves

- Year 1
  - Day 1
  - Day 2
    <!-- NOW -->
    - Danced
`

func TestPivotByGroup(t *testing.T) {
	var output bytes.Buffer
	if err := pivotByGroup(strings.NewReader(byGroupInput), &output); err != nil {
		t.Fatalf("pivotByGroup returned error: %v", err)
	}
	if got := output.String(); got != byGroupExpected {
		t.Errorf("pivotByGroup output mismatch\ngot:\n%s\nwant:\n%s", got, byGroupExpected)
	}
}
