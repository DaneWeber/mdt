/*
Copyright © 2026 Dane Weber
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
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

func pivot(r io.Reader, w io.Writer) error {
	type block struct {
		section    string
		subSection string
		lines      []string
	}

	var blocks []block
	var sectionOrder []string
	var subSectionOrder []string
	seenSections := map[string]bool{}
	seenSubSections := map[string]bool{}

	var curSection, curSubSection string
	var curLines []string

	flush := func() {
		if curSection == "" || curSubSection == "" {
			return
		}
		lines := curLines
		// trim leading blank lines
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
		// trim trailing blank lines
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		blocks = append(blocks, block{curSection, curSubSection, lines})
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## "):
			flush()
			curSection = strings.TrimPrefix(line, "# ")
			curSubSection = ""
			curLines = nil
			if !seenSections[curSection] {
				seenSections[curSection] = true
				sectionOrder = append(sectionOrder, curSection)
			}
		case strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### "):
			flush()
			curSubSection = strings.TrimPrefix(line, "## ")
			curLines = nil
			if !seenSubSections[curSubSection] {
				seenSubSections[curSubSection] = true
				subSectionOrder = append(subSectionOrder, curSubSection)
			}
		default:
			if curSubSection != "" {
				curLines = append(curLines, line)
			}
		}
	}
	flush()

	if err := scanner.Err(); err != nil {
		return err
	}

	type key struct{ sub, sec string }
	data := map[key][]string{}
	for _, b := range blocks {
		data[key{b.subSection, b.section}] = b.lines
	}

	for i, sub := range subSectionOrder {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s\n", sub)
		for _, sec := range sectionOrder {
			lines, ok := data[key{sub, sec}]
			if !ok {
				continue
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "### %s\n", sec)
			fmt.Fprintln(w)
			for _, l := range lines {
				fmt.Fprintln(w, l)
			}
		}
	}

	return nil
}
