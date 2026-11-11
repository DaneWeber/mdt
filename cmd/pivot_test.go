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
