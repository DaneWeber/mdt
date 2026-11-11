package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func runLevels(t *testing.T, input string, levels, skip int) string {
	t.Helper()
	var out, warn bytes.Buffer
	if err := pivotLevels(strings.NewReader(input), &out, levels, skip, &warn); err != nil {
		t.Fatalf("pivotLevels returned error: %v", err)
	}
	if warn.Len() > 0 {
		t.Errorf("unexpected warnings:\n%s", warn.String())
	}
	return out.String()
}

const byDate = `# Log

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

const byGroup = `# Log

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

func TestPivotLevelsForward(t *testing.T) {
	if got := runLevels(t, byDate, 1, 1); got != byGroup {
		t.Errorf("mismatch\ngot:\n%s\nwant:\n%s", got, byGroup)
	}
}

func TestPivotLevelsReverse(t *testing.T) {
	if got := runLevels(t, byGroup, 2, 1); got != byDate {
		t.Errorf("mismatch\ngot:\n%s\nwant:\n%s", got, byDate)
	}
}

func TestPivotLevelsSkip(t *testing.T) {
	input := `# Campaigns

## Campaign A

### Day 1

- Dwarves
  - Dug

## Campaign B

### Day 1

- Elves
  - Sang
`
	want := `# Campaigns

## Campaign A

### Dwarves

- Day 1
  - Dug

## Campaign B

### Elves

- Day 1
  - Sang
`
	if got := runLevels(t, input, 1, 2); got != want {
		t.Errorf("mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestPivotLevelsSampleRoundTrip(t *testing.T) {
	orig, err := os.ReadFile("../sample-files/background-activity.dm.md")
	if err != nil {
		t.Fatal(err)
	}
	grouped := runLevels(t, string(orig), 1, 1)
	if back := runLevels(t, grouped, 3, 1); back != string(orig) {
		t.Errorf("round trip mismatch\ngot:\n%s\nwant:\n%s", back, orig)
	}
}

func TestPivotLevelsWarnsOnDroppedContent(t *testing.T) {
	var out, warn bytes.Buffer
	input := "# T\n\n## Day 1\n\nA stray paragraph.\n\n- Dwarves\n"
	if err := pivotLevels(strings.NewReader(input), &out, 1, 1, &warn); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warn.String(), "A stray paragraph.") {
		t.Errorf("expected warning about dropped paragraph, got %q", warn.String())
	}
}
