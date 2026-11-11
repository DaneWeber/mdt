/*
Copyright © 2026 Dane Weber
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var (
	pivotOutput  string
	pivotLevelsN int
	pivotSkip    int
)

// pivotCmd represents the pivot command
var pivotCmd = &cobra.Command{
	Use:   "pivot [flags] input.md...",
	Short: "Pivot the heading structure of Markdown files",
	Long: `Pivot the structure of one or more Markdown files.

By default, H1 sections containing H2 subsections are swapped so that each
H2 becomes a top-level section containing the matching content from every H1.

With --levels N, the first N levels of the lists become headings and the
heading levels become nested list items. The top --skip heading levels (default
1, e.g. the document title) are kept as they are. Running it again with
--levels set to the number of former heading levels reverses it.

Example: a file with "## Year / ### Month / #### Day" headings over lists of
groups becomes one "## Group" section per group with -l 1, and turns back
with -l 3.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputs []io.Reader
		for _, name := range args {
			f, err := os.Open(name)
			if err != nil {
				return err
			}
			defer f.Close()
			inputs = append(inputs, f, strings.NewReader("\n"))
		}
		in := io.MultiReader(inputs...)

		var out bytes.Buffer
		var err error
		if cmd.Flags().Changed("levels") {
			err = pivotLevels(in, &out, pivotLevelsN, pivotSkip, os.Stderr)
		} else {
			err = pivot(in, &out)
		}
		if err != nil {
			return err
		}

		if pivotOutput == "" || pivotOutput == "-" {
			_, err := os.Stdout.Write(out.Bytes())
			return err
		}
		return os.WriteFile(pivotOutput, out.Bytes(), 0o644)
	},
}

func init() {
	rootCmd.AddCommand(pivotCmd)
	pivotCmd.Flags().StringVarP(&pivotOutput, "output", "o", "", "output file (default stdout)")
	pivotCmd.Flags().IntVarP(&pivotLevelsN, "levels", "l", 0, "number of list levels to turn into headings; headings become list items (omit for the plain H1/H2 swap)")
	pivotCmd.Flags().IntVarP(&pivotSkip, "skip", "s", 1, "number of top heading levels to keep unchanged (used with --levels)")
}

// headingText extracts plain text from a heading's inline children.
func headingText(h *gast.Heading, src []byte) string {
	var sb strings.Builder
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*gast.Text); ok {
			sb.Write(t.Segment.Value(src))
		}
	}
	return strings.TrimSpace(sb.String())
}

// contentSource returns the raw source markdown for a block node.
// Segments stored in leaf nodes begin after list markers, so we back up
// to the true start of each line to include the marker and indentation.
func contentSource(n gast.Node, src []byte) string {
	start := len(src)
	end := 0
	_ = gast.Walk(n, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if node.Type() == gast.TypeInline {
			if entering {
				return gast.WalkSkipChildren, nil
			}
			return gast.WalkContinue, nil
		}
		if !entering {
			return gast.WalkContinue, nil
		}
		lines := node.Lines()
		if lines == nil || lines.Len() == 0 {
			return gast.WalkContinue, nil
		}
		first := lines.At(0)
		// back up to the real start of the line (before list markers, etc.)
		ls := first.Start
		for ls > 0 && src[ls-1] != '\n' {
			ls--
		}
		if ls < start {
			start = ls
		}
		if last := lines.At(lines.Len() - 1); last.Stop > end {
			end = last.Stop
		}
		return gast.WalkContinue, nil
	})
	if start >= end {
		return ""
	}
	return strings.TrimRight(string(src[start:end]), "\n")
}

func pivot(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	doc := goldmark.New().Parser().Parse(text.NewReader(src))

	type key struct{ sub, sec string }
	var sectionOrder, subSectionOrder []string
	seenSections := map[string]bool{}
	seenSubSections := map[string]bool{}
	data := map[key][]string{}
	var curSection, curSubSection string

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*gast.Heading); ok {
			txt := headingText(h, src)
			switch h.Level {
			case 1:
				curSection = txt
				curSubSection = ""
				if !seenSections[curSection] {
					seenSections[curSection] = true
					sectionOrder = append(sectionOrder, curSection)
				}
			case 2:
				curSubSection = txt
				if !seenSubSections[curSubSection] {
					seenSubSections[curSubSection] = true
					subSectionOrder = append(subSectionOrder, curSubSection)
				}
			}
		} else if curSection != "" && curSubSection != "" {
			k := key{curSubSection, curSection}
			if content := contentSource(n, src); content != "" {
				data[k] = append(data[k], content)
			}
		}
	}

	for i, sub := range subSectionOrder {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s\n", sub)
		for _, sec := range sectionOrder {
			contents, ok := data[key{sub, sec}]
			if !ok {
				continue
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "### %s\n", sec)
			fmt.Fprintln(w)
			fmt.Fprintln(w, strings.Join(contents, "\n\n"))
		}
	}

	return nil
}
