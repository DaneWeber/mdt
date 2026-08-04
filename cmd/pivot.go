/*
Copyright © 2026 Dane Weber
*/
package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// pivotCmd represents the pivot command
var pivotCmd = &cobra.Command{
	Use:   "pivot",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pivot called")
	},
}

func init() {
	rootCmd.AddCommand(pivotCmd)
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
